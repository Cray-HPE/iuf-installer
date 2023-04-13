---
date: 2023-04-13T10:14:46-04:00
title: "iuf-installer completion bash"
layout: default
---
## iuf-installer completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(iuf-installer completion bash)

To load completions for every new session, execute once:

#### Linux:

	iuf-installer completion bash > /etc/bash_completion.d/iuf-installer

#### macOS:

	iuf-installer completion bash > $(brew --prefix)/etc/bash_completion.d/iuf-installer

You will need to start a new shell for this setup to take effect.


```
iuf-installer completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [iuf-installer completion](/commands/iuf-installer_completion/)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 13-Apr-2023