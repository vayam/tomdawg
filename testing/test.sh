curl -Fs=@tomdawg http://localhost:8089/short
echo ""
curl -X PUT -H "Content-Type: video/mp4" --data-binary @tomdawg http://localhost:8089/short
echo ""
curl -X PUT --data-binary @video.mp4 http://localhost:8089/put/test/video.mp4
echo ""
curl -F multipart/test=@video.mp4 http://localhost:8089
echo ""
curl -X PUT --data-binary @tomdawg http://localhost:8089/put/tomdawg
echo ""
curl -F multipart/=@tomdawg http://localhost:8089
echo ""
