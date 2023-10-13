# #!/bin/bash

# set -e

# # wait for MinIO to start
# sleep 5

# sudo curl -o /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc

# sudo chmod +x /usr/local/bin/mc
# mc --version

# # create the bucket
# mc alias set myminio http://localhost:9000 masoud Strong#Pass#2022
# mc ls myminio
# mc mb myminio/vilow-videos

# # set the bucket policy
# mc policy set download myminio/vilow-videos

##### ------------------------------

#!/bin/sh
set -e

# wait for MinIO to start
sleep 10

apk update
apk add curl

# Update the MinIO URL to point to the correct container name
MINIO_URL="http://minio:9000"
MINIO_ACCESS_KEY="$MINIO_ROOT_USER"
MINIO_SECRET_KEY="$MINIO_ROOT_PASSWORD"

curl -o /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x /usr/local/bin/mc
mc --version

mc alias set myminio $MINIO_URL $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
mc ls myminio
mc mb myminio/vilow-videos

mc admin policy create myminio read-write-policy /scripts/policies/read-write-policy.json

mc policy public myminio/vilow-videos

mc policy set myminio/vilow-videos read-write-policy
mc policy set upload myminio/vilow-videos
mc policy set download myminio/vilow-videos

echo "bucket config done."


