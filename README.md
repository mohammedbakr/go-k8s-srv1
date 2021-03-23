# icap-service-1

This is the service 1 of [this repo](https://github.com/k8-proxy/go-k8s-infra) that will get the file from icap-server and upload it to minio

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
git clone https://github.com/k8-proxy/go-k8s-srv1.git
cd k8-proxy/go-k8s-srv1
docker build -t <docker_image_name> .
```

- To run the container
First make sure that you have rabbitmq and minio running, then run the command bellow 

```
docker run -e ADAPTATION_REQUEST_QUEUE_HOSTNAME='<rabbit-host>' \ 
-e ADAPTATION_REQUEST_QUEUE_PORT='<rabbit-port>' \
-e MESSAGE_BROKER_USER='<rabbit-user>' \
-e MESSAGE_BROKER_PASSWORD='<rabbit-password>' \
-e MINIO_ENDPOINT='<minio-endpoint>' \ 
-e MINIO_ACCESS_KEY='<minio-access>' \ 
-e MINIO_SECRET_KEY='<minio-secret>' \ 
-e MINIO_SOURCE_BUCKET='<bucket-to-upload-file>' \ 
--name <docker_container_name> <docker_image_name>
```

# Testing steps

- Run the container as mentionned above

- Publish data reference to rabbitMq on queue name : adaptation-request-queue with the following data(table) :
* file-id : An ID for the file
* source-file-location : The full path to the file
* rebuilt-file-location : A full path representing the location where the rebuilt file will go to


- Check your container logs to see the processing

```
docker logs <container name>
```

# Rebuild flow to implement

![new-rebuild-flow-v2](https://github.com/k8-proxy/go-k8s-infra/raw/main/diagram/go-k8s-infra.png)
