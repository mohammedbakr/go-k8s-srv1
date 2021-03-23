# icap-service-1

This service is as per the architecture the first proxy service that will get the file from icap-server and upload it to minio

### Steps of processing
When it starts
- Listens on the queue
- Get the file from the queue
- upload it to minio bucket
- Push the file details to the queue

## Configuration
- This pod need to mount the share storage mounted on icap server and that is how they will share the file together
- It's possible to have multiple replica of this service running. Only one will get the file and process it


### Docker build
- To build the docker image
```
docker build -t <docker_image_name> .
```

# Testing steps

- Log in to the VM
- Make sure that all the pods are running

```
kubectl  -n icap-adaptation get pods
```

- Start a test using the command bellow : If all is ok you will receive a result file.

```
mkdir /tmp/input
cp <pdf_file_name> /tmp/input/
docker run --rm -v /tmp/input:/opt/input -v /tmp/output:/opt/output glasswallsolutions/c-icap-client:manual-v1 -s 'gw_rebuild' -i <your vm IP> -f '/opt/input/<pdf_file_name>' -o /opt/output/<pdf_file_name> -v
```

During the test check logs of icap-service1 pods, they should have get and processed the file

# Rebuild flow to implement

![new-rebuild-flow-v2](https://user-images.githubusercontent.com/76431508/107766490-35064200-6d3c-11eb-8d63-ad64f29ce964.jpeg)
