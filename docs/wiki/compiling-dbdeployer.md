# Compiling dbdeployer
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Should you need to compile your own binaries for dbdeployer, follow these steps:

1. Make sure you have go 1.11+ installed in your system.
2. Run `git clone https://github.com/ProxySQL/dbdeployer.git`.  This will import all the code that is needed to build dbdeployer.
3. Change directory to `./dbdeployer`.
4. Run ./scripts/build.sh {linux|OSX}`
5. If you need the docs enabled binaries (see the section "Generating additional documentation") run `MKDOCS=1 ./scripts/build.sh {linux|OSX}`

