curl -Fs=@../tomdawg http://localhost:8089/short
echo ""
rm -rf ../files/*
curl -X PUT -H "Content-Type: video/mp4" --data-binary @../tomdawg http://localhost:8089/short
echo ""
rm -rf ../files/*
curl -X PUT --data-binary @video.mp4 http://localhost:8089/put/test/video.mp4
echo ""
rm -rf ../files/*
curl -F multipart/test=@video.mp4 http://localhost:8089
echo ""
rm -rf ../files/*
curl -X PUT --data-binary @../tomdawg http://localhost:8089/put/../tomdawg
echo ""
rm -rf ../files/*
curl -F multipart/=@../tomdawg http://localhost:8089
echo ""
rm -rf ../files/*
echo "foo" | curl -X PUT -H "Transfer-Encoding: chunked" -d @- http://localhost:8089/foo -v
echo ""
rm -rf ../files/*
cat ../tomdawg | curl -X PUT -H "Transfer-Encoding: chunked" --data-binary @- http://localhost:8089/tomdawg -v
echo ""
rm -rf ../files/*

