To know how to install requirements of gRPC, refer to the link [https://grpc.io/docs/languages/go/quickstart/](https://grpc.io/docs/languages/go/quickstart/).
`.proto` files are stored in the `pkg/pb` (pb: Protocol Buffer) directory.  
If you edited one of the `.proto` files, you must regenerate code with executing below code from the root directory.
For regenerating auth-service code, execute:
```sh
protoc --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       pkg/pb/auth/auth-service.proto
```

HTTP cpde responses:
`500`: Internal Server Error
`400`: Bad Request. e.g.  bad request structure
`401`: Unauthorized
`403`: Forbidden
TODO: Create an error if the file size in uploading is greater than the limit

For downloading a file, objectToken, is the name of the file (that is stored in the storage). Also, if corresponding URL of a token is empty, means that the file couldn't be downloaded by the user.

In HTTP response of a file uploading request, if value of
a extension value be empty, means you are not allowed to upload the file.

We default download and upload links are `/download` and `/upload`.

*How to upload a file?*
1) Create a HTTP POST request that specifies number of each file type/extension you're going to upload along with authentication token. Curl command for this request:
```sh
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{"auth-token": "token", "object-types":{
      "jpg": 10,
      "pdf": 1
     } \
    }' \
     http://API_URL/upload
```

TODO: Add these features: Set maximum upload size (of a file) - Assign/read labels/metadata to files

**How to create docker image for the app:**
1) Create a docker image for the app:  
```
docker build -t file-transfer .
```

**How to run this in Kubernetes:**
1) Follow instructions in the [README.md](https://github.com/q-sharafian/DMS/blob/master/README.md) file.