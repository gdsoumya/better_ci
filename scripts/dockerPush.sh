#echo $1 $2 $3
echo $2 | docker login --username $1 --password-stdin && docker push $3
