@for /f %%i in ('git.exe rev-parse --short HEAD') do set REVISION=%%i
packr build -o thriftkit.exe -v -ldflags "-X github.com/jxskiss/thriftkit/generator.GitRevision=%REVISION%"
