# sweb
Swagger Editor Backend

## Installation

```
go get -u github.com/oRRs/sweb
```

## usage

```
Usage of sweb:
  -f string
    	the full path to the document being edited (default "api-spec.yaml")
  -p string
    	port for editor's http backend (default "8765")
  -se string
    	the full path to swagger-editor installation (default "builtin")
  -u string
      username for web authentication (default "admin")
  -k string
      password for web authentication (default "admin")
```

### Important stuff

The backend can only run in one instance, and you'll be able to work on one file at a time. If there will be demand, it is possible that it will support multiple files.
There is no save button. Whatever you do is saved. There is undo in the editor, but it's still easy to mess up. The idea is that you use this tool inside a git repository
so you can revert changes etc. 
