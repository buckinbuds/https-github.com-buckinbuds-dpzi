https://chainlinkadapters.com/details/smartcontractkit/paypal-adapter
PayPal External Adapter
**THIS WILL NEED TO BE MODIFIED PRIOR IMPLEMENTATION**
**create zip file to upload to AWS/GCP, run:**
zip -r cl-ea.zip.
</br>
</br>
**to run with docker, use this code:**
docker build . -t paypal-adapter
docker run -d \
    -p 8080:8080 \
    -e EA_PORT=8080 \
    -e CLIENT_ID="Your_client_id" \
    -e CLIENT_SECRET="Your_client_secret" \
    paypal-adapter
