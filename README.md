tomdawg
=======

simple upload server


##build
go build tomdawg.go

##configure
edit config.json

```json
{
    "ListenPort": 8089,
    "AssetPath": "./files"
}
```

##run server
./tomdawg

##logs
you can find log file here:  ./logs/tomdawg.log

##usage
with PUT, the path is determined from url path
```
curl -X PUT --data-binary @video.mp4 http://localhost:8089/put/test/video.mp4
```
sample response:
```json

{
"Path":"/Users/vayam/vayam-dev/tomdawg/files/put/test/video.mp4",
"Status":"success",
"Description":"Uploaded successfully",
"Time":0,
"Speed":0,
"Size":1127145,
"Recvd":1127145
}
```

with regular POST, path is determined by upload form name,value
```
curl -F multipart/test=@video.mp4 http://localhost:8089
```
sample response:
```json

{
"Path":"/Users/vayam/vayam-dev/tomdawg/files/multipart/test/video.mp4",
"Status":"success",
"Description":"Total Files: 1 Total Bytes: 1127145",
"Time":0,
"Speed":0,
"Size":0,
"Size":1127145,
"Recvd":1127145
}
```
