## example

### Профиль cpu
> go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/profile

### Память
> go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/heap

## codeprofile
> curl localhost:8080/profile 
> go tool pprof -http=":9090" mem.profile
> go tool pprof -http=":9090" cpu.profile

## mutex
[pprof](http://127.0.0.1:8081/debug/pprof/)

## labels
> curl localhost:8080/profile/1
> go tool pprof -seconds=30 http://localhost:8080/debug/pprof/profile
> tags
> tagfocus=handler:1
> top
> tagfocus=
> tagignore=handler:1
> top
> web

## trace
> curl "http://localhost:8080/debug/pprof/trace?seconds=15" > trace.out
> go tool trace -http='localhost:9090' trace.out
