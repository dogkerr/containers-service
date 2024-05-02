






#### Cara pake hertz & kitex framework

Hertz:
```
1. bikin idl (<nama_protobuffile>.proto)
2. hz new -module dogker/lintang/container-service -idl idl/<nama_protobuffile>.proto
3. go mod tidy
4.. buat update /tambahin endpoint tinggal tambahin idl  (<nama_protobuffile>.proto)
5. hz  -module dogker/lintang/container-service  update -idl idl/hello.proto 
```

Kitex:
```
1. bikin idl (grpc/<nama_protobuffile>.proto)
2. kitex -module dogker/lintang/container-service -I idl/  -type protobuf     --protobuf Mgoogle/protobuf/descriptor.proto=A-Import-Path-In-kitex_gen     idl/grpc/<nama_protobuffile>.proto

3. kitex -type protobuf  -module dogker/lintang/container-service -service hello.service  -I ./idl/  ./idl/grpc/<nama_protofbuffile>.proto


import (
	pb "dogker/lintang/container-service/kitex_gen/container-service/pb/containerservice"
	"log"
	"net"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/kitex/pkg/transmeta"
	kitexServer "github.com/cloudwego/kitex/server"
)

func main() {
	h := server.Default()
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:6000")
	var opts []kitexServer.Option
	opts = append(opts, kitexServer.WithMetaHandler(transmeta.ServerHTTP2Handler))
	opts = append(opts, kitexServer.WithServiceAddr(addr))
	svr := pb.NewServer(new(ContainerServiceImpl), opts...) // kitex rpc server
	go func() {
		err := svr.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	register(h)
	h.Spin()
}

4. go mod tidy

```


#### cara run
```

sh build.sh
 sh output/bootstrap.sh
 ```



