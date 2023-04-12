#!/usr/bin/env bash
#
# MIT License
#
# (C) Copyright 2022 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
set -euo pipefail
set -o xtrace

: "${RELEASE:="${RELEASE_NAME:="csm"}-${RELEASE_VERSION:="0.0.0"}"}"

# Define maximums for retries on skopeo i/o timeout bandaid logic
export MAX_SKOPEO_RETRY_ATTEMPTS=20
export MAX_SKOPEO_RETRY_TIME_MINUTES=30

# import release utilities
ROOTDIR="$(dirname "${BASH_SOURCE[0]}")"
source "${ROOTDIR}/vendor/github.hpe.com/hpe/hpc-shastarelm-release/lib/release.sh"

requires curl docker git perl rsync sed

# Valid SemVer regex, see https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
semver_regex='^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'

# Release version must be a valid semver
if [[ -z "$(echo "$RELEASE_VERSION" | perl -ne "print if /$semver_regex/")" ]]; then
    echo >&2 "error: invalid RELEASE_VERSION: ${RELEASE_VERSION}"
    exit
fi

# Parse components of version
RELEASE_VERSION_MAJOR="$(echo "$RELEASE_VERSION" | perl -pe "s/${semver_regex}/\1/")"
RELEASE_VERSION_MINOR="$(echo "$RELEASE_VERSION" | perl -pe "s/${semver_regex}/\2/")"
RELEASE_VERSION_PATCH="$(echo "$RELEASE_VERSION" | perl -pe "s/${semver_regex}/\3/")"
RELEASE_VERSION_PRERELEASE="$(echo "$RELEASE_VERSION" | perl -pe "s/${semver_regex}/\4/")"
RELEASE_VERSION_BUILDMETADATA="$(echo "$RELEASE_VERSION" | perl -pe "s/${semver_regex}/\5/")"

#
# Setup
#

#code to store credentials in environment variable
if [ ! -z "$ARTIFACTORY_USER" ] && [ ! -z "$ARTIFACTORY_TOKEN" ]; then
    export REPOCREDSVARNAME="REPOCREDSVAR"
    export REPOCREDSVAR=$(jq --null-input --arg url "https://artifactory.algol60.net/artifactory/" --arg realm "Artifactory Realm" --arg user "$ARTIFACTORY_USER"   --arg password "$ARTIFACTORY_TOKEN"   '{($url): {"realm": $realm, "user": $user, "password": $password}}')
fi

# Load and verify assets
source "${ROOTDIR}/assets.sh"

# Build image list (and sync charts to build/.helm/cache/repository
make -C "$ROOTDIR" images

# Pull release tools
cmd_retry docker pull "$PACKAGING_TOOLS_IMAGE"
cmd_retry docker pull "$RPM_TOOLS_IMAGE"
cmd_retry docker pull "$SKOPEO_IMAGE"
cmd_retry docker pull "$CRAY_NEXUS_SETUP_IMAGE"

#
# Build
#

BUILDDIR="${1:-"$(realpath -m "$ROOTDIR/dist/${RELEASE}")"}"

# Initialize build directory
[[ -d "$BUILDDIR" ]] && rm -fr "$BUILDDIR"
mkdir -p "$BUILDDIR"

# Copy install scripts
# rsync -aq "${ROOTDIR}/lib/" "${BUILDDIR}/lib/"
# gen-version-sh "$RELEASE_NAME" "$RELEASE_VERSION" >"${BUILDDIR}/lib/version.sh"
# chmod +x "${BUILDDIR}/lib/version.sh"
# rsync -aq "${ROOTDIR}/vendor/github.hpe.com/hpe/hpc-shastarelm-release/lib/install.sh" "${BUILDDIR}/lib/install.sh"
# rsync -aq "${ROOTDIR}/install.sh" "${BUILDDIR}/"
# rsync -aq "${ROOTDIR}/chrony/" "${BUILDDIR}/chrony/"
# rsync -aq "${ROOTDIR}/upgrade.sh" "${BUILDDIR}/"
# rsync -aq "${ROOTDIR}/hack/load-container-image.sh" "${BUILDDIR}/hack/"

# Copy manifests
rsync -aq "${ROOTDIR}/manifests/" "${BUILDDIR}/manifests/"

# Configure yq
shopt -s expand_aliases
alias yq="${ROOTDIR}/vendor/stash.us.cray.com/scm/shasta-cfg/stable/utils/bin/$(uname | awk '{print tolower($0)}')/yq"

# Rewrite manifest spec.sources.charts to reference local helm directory
find "${BUILDDIR}/manifests/" -name '*.yaml' | while read manifest; do
    yq w -i -s - "$manifest" << EOF
- command: update
  path: spec.sources.charts[*].type
  value: directory
- command: update
  path: spec.sources.charts[*].location
  value: ./helm
- command: delete
  path: spec.sources.charts[*].credentialsSecret
EOF
done

# Embed the CSM release version into the csm-config and cray-csm-barebones-recipe-install charts
# yq write -i ${BUILDDIR}/manifests/sysmgmt.yaml 'spec.charts.(name==csm-config).values.cray-import-config.import_job.CF_IMPORT_PRODUCT_NAME' "$RELEASE_NAME"
# yq write -i ${BUILDDIR}/manifests/sysmgmt.yaml 'spec.charts.(name==csm-config).values.cray-import-config.import_job.CF_IMPORT_PRODUCT_VERSION' "$RELEASE_VERSION"
# yq write -i ${BUILDDIR}/manifests/sysmgmt.yaml 'spec.charts.(name==csm-config).values.cray-import-config.import_job.CF_IMPORT_GITEA_REPO' "${RELEASE_NAME}-config-management"
# yq write -i ${BUILDDIR}/manifests/sysmgmt.yaml 'spec.charts.(name==cray-csm-barebones-recipe-install).values.cray-import-kiwi-recipe-image.import_job.PRODUCT_VERSION' "${RELEASE_VERSION}"
# yq write -i ${BUILDDIR}/manifests/sysmgmt.yaml 'spec.charts.(name==cray-csm-barebones-recipe-install).values.cray-import-kiwi-recipe-image.import_job.PRODUCT_NAME' "${RELEASE_NAME}"
# yq write -i ${BUILDDIR}/manifests/sysmgmt.yaml 'spec.charts.(name==cray-csm-barebones-recipe-install).values.cray-import-kiwi-recipe-image.import_job.name' "${RELEASE_NAME}-image-recipe-import-${RELEASE_VERSION}"

# Include the tds yaml in the tarball
# cp "${ROOTDIR}/tds_cpu_requests.yaml" "${BUILDDIR}/tds_cpu_requests.yaml"

# Generate Nexus blob store configuration
# generate-nexus-config blobstore <"${ROOTDIR}/nexus-blobstores.yaml" >"${BUILDDIR}/nexus-blobstores.yaml"

# Generate Nexus repositories configuration
# Update repository names based on the release version
# sed -e "s/-0.0.0/-${RELEASE_VERSION}/g" "${ROOTDIR}/nexus-repositories.yaml" \
#     | generate-nexus-config repository >"${BUILDDIR}/nexus-repositories.yaml"

# # Sync shasta-cfg
# mkdir "${BUILDDIR}/shasta-cfg"
# "${ROOTDIR}/vendor/stash.us.cray.com/scm/shasta-cfg/stable/package/make-dist.sh" "${BUILDDIR}/shasta-cfg"

# Sync Helm charts from cache
rsync -aq "${ROOTDIR}/build/.helm/cache/repository"/*.tgz "${BUILDDIR}/helm"

# Sync container images
parallel -j 75% --retries 5 --halt-on-error now,fail=1 -v \
    -a "${ROOTDIR}/build/images/index.txt" --colsep '\t' \
    "${ROOTDIR}/build/images/sync.sh" "docker://{2}" "dir:${BUILDDIR}/docker/{1}"

# Sync RPM manifests
export RPM_SYNC_NUM_CONCURRENT_DOWNLOADS=32
rpm-sync "${ROOTDIR}/rpm/cray/csm/sle-15sp4/index.yaml" "${BUILDDIR}/rpm/cray/csm/sle-15sp4" -s

# Fix-up cray directories by removing misc subdirectories
{
    find "${BUILDDIR}/rpm/cray" -name '*-team' -type d
    find "${BUILDDIR}/rpm/cray" -name 'github' -type d
} | while read path; do
    mv "$path"/* "$(dirname "$path")/"
    rmdir "$path"
done

# Remove empty directories
find "${BUILDDIR}/rpm/cray" -empty -type d -delete

# Create CSM repositories
createrepo "${BUILDDIR}/rpm/cray/csm/sle-15sp4"

# Clean up temp space
rm -fr "${BUILDDIR}/tmp"


# Download HPE GPG signing key (for verifying signed RPMs)
cmd_retry curl -sfSLRo "${BUILDDIR}/hpe-signing-key.asc" "$HPE_SIGNING_KEY"

# Save cray/nexus-setup and quay.io/skopeo/stable images for use in install.sh
vendor-install-deps "$(basename "$BUILDDIR")" "${BUILDDIR}/vendor"

# Package the distribution into an archive
tar -C "${BUILDDIR}/.." --owner=0 --group=0 -cvzf "${BUILDDIR}/../$(basename "$BUILDDIR").tar.gz" "$(basename "$BUILDDIR")/" --remove-files
