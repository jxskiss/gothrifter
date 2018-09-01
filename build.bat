@for /f %%i in ('git.exe rev-parse --short HEAD') do set REVISION=%%i
packr build -o thrifterc.exe -v -ldflags "-X github.com/jxskiss/gothrifter/generator.GitRevision=%REVISION%"
