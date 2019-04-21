# learning-kafka

#### start kafka cluster
docker-compose --file ./cluster.yaml up -d

#### install libkafka driver
1. get and start cygwin64 installation from https://www.cygwin.com/setup-x86_64.exe
1. install `x86_64-w64-mingw32-gcc` and `pkg-config`
1. add `c:\cygwin64\bin` to PATH
1. et nuget from https://dist.nuget.org/win-x86-commandline/latest/nuget.exe
1. run nuget install librdkafka.redist -Version 1.0.0 ( this is currently the latest version )
1. copy `.\librdkafka.redist.0.11.4\build\native\include\` into `c:\cygwin64\usr\include\`
1. copy `.\librdkafka.redist.0.11.4\build\native\lib\win7\x64\win7-x64-Release\v120\librdkafka.lib` into `c:\cygwin64\lib\librdkafka.a` (notice .lib is renamed to .a)
1. copy `librdkafka.dll`, `librdkafkacpp.dll`, `libzstd.dll` and `zlib.dll` from `.\librdkafka.redist.0.11.4\runtimes\win7-x64\native\` to `C:\Windows\System32`
1. create file `rdkafka.pc` into `c:\cygwin64\bin` with following content:
    ```
    prefix=c:/
    libdir=c:/cygwin64/lib/
    includedir=c:/cygwin64/usr/include

    Name: librdkafka
    Description: The Apache Kafka C/C++ library
    Version: 0.11.4
    Cflags: -I${includedir}
    Libs: -L${libdir} -lrdkafka
    Libs.private: -lssl -lcrypto -lcrypto -lz -ldl -lpthread -lrt
    ```
1. run `set CC=x86_64-w64-mingw32-gcc` (this will allow cgo to use x86_64-w64-mingw32-gcc instead of gcc - you can make sure it worked with `go env CC`)
1. run go build, that will create executable

#### reference:
https://github.com/confluentinc/confluent-kafka-go/issues/128
https://github.com/simplesteph/kafka-stack-docker-compose
https://3gods.com/bigdata/Kafka-Message-Delivery-Semantics.html#sec-1-4
https://blog.csdn.net/qq_20641565/article/details/64440425

https://cutejaneii.wordpress.com/2017/11/22/kafka-6-%E8%A8%AD%E5%AE%9Atopic%E7%9A%84%E5%89%AF%E6%9C%AC%E6%95%B8replication-factor-%E5%88%A9%E7%94%A8server-properties%E9%A0%90%E8%A8%AD%E6%95%B8%E9%87%8F-vs-%E5%BB%BA%E7%AB%8Btopic/

#### Tool:
http://www.kafkatool.com/features.html