# Service-MQTT

This service manages MQTT-enabled devices. Its primary functions are:

1.  **Device Management**: Dynamically provisions new devices with unique MQTT credentials via an HTTP API.
2.  **Authentication**: Authenticates devices connecting to the MQTT broker against credentials stored in a PostgreSQL database.
3.  **Data Forwarding**: Forwards all messages published by authenticated devices to a NATS cluster for further processing.

The service is designed to be scalable and is deployable on Kubernetes.

-----

## System Architecture

The diagram below illustrates the flow of data and interactions between the different components of the system.

<img alt="image" src="https://github.com/user-attachments/assets/841db167-89d7-4a45-8371-f8a7ad2065b3" />

**Flow Description**:

1.  An external client sends an **"Add Device"** request to the HTTP API endpoint to provision a new device.
2.  The **Device Manager** receives the request, generates new credentials, and saves the device information to the **DB** (PostgreSQL).
3.  A **Device** connects to the MQTT **Broker** using its provisioned credentials.
4.  The **Broker** validates the device's credentials with the **Device Manager**, which queries the **DB**.
5.  Once authenticated, the device publishes MQTT messages to the **Broker**.
6.  The **Broker** forwards the messages to the **Publisher**, which then publishes the data to a **NATS** stream for other services to consume.

-----

## Core Technologies

This service is built with Go and leverages several key technologies:

* **Go**: The primary programming language.
* **Docker**: The application is containerized for easy deployment and scaling.
* **PostgreSQL**: Used as the database to store device credentials and information.
* **MQTT Broker**: A lightweight messaging protocol for small sensors and mobile devices. The service uses the `mochi-mqtt/server` library.
* **NATS**: A high-performance messaging system used to forward data from the MQTT broker for further processing.
* **Kubernetes**: The service is designed to be deployed on a Kubernetes cluster, with manifest files provided in the `deploy/` directory.

-----

## API

The service exposes an HTTP API for managing devices.

### Add a New Device

Creates a new device and returns its generated MQTT credentials.

* **Endpoint**: `POST /devices`
* **Request Body**:
  ```json
  {
      "type": "string"
  }
  ```
* **Success Response (200 OK)**:
  ```json
  {
      "id": "string",
      "type": "string",
      "mqtt_user": "string",
      "mqtt_password": "string",
      "created_at": "string"
  }
  ```

This API specification is detailed in the Swagger documentation available at the `/docs/index.html` endpoint.

-----

## How to Run Locally

You can run the entire system locally using Docker Compose.

1.  **Prerequisites**:

    * Docker and Docker Compose installed.

2.  **Run the service**:
    From the root of the project, run the following command:

    ```sh
    docker-compose up
    ```

    This will start the `service-mqtt`, a PostgreSQL database, and a NATS server.

    * The MQTT broker will be available on port `1883`.
    * The HTTP API will be available on port `9091`.

-----

## Publishing a Message

Once the service is running, you can test it by provisioning a device and publishing a message with it.

1.  **Provision a New Device**

    Open a new terminal and use `curl` to send a request to the API to create a new device. This will return the credentials needed to connect to the MQTT broker.

    ```sh
    curl -X POST http://localhost:9091/devices -d '{"type":"test-device"}'
    ```

    You will get a JSON response with an `mqtt_user` (which is the same as the `id`) and an `mqtt_password`.

2.  **Publish a Message**

    Use the `eclipse-mosquitto` Docker image to publish a message. Replace `<YOUR_USERNAME>` and `<YOUR_PASSWORD>` with the credentials you received from the API call. The topic should be in the format `raw.<YOUR_USERNAME>`.

    ```sh
    docker run --rm -it --network=host eclipse-mosquitto mosquitto_pub \
    -h localhost \
    -p 1883 \
    -u "<YOUR_USERNAME>" \
    -P "<YOUR_PASSWORD>" \
    -t "raw.<YOUR_USERNAME>" \
    -m "Hello from my new device!"
    ```

    You should see the message being forwarded to NATS in the logs from the `docker-compose up` command.

-----

## Deployment

The YAML files required to deploy this service on a Kubernetes cluster are located in the `deploy/` directory. These manifests define the necessary resources, including:

* **Namespace**: Creates a `scadable-io` namespace for the application.
* **PersistentVolumeClaim**: For PostgreSQL data storage.
* **PostgreSQL StatefulSet**: Deploys the database.
* **Deployment**: Manages the `service-mqtt` application pods.
* **Service**: Exposes the application's HTTP and MQTT ports within the cluster.
* **Ingress**: Manages external access to the HTTP API.
* **ConfigMap**: Configures TCP service passthrough for the MQTT broker via an NGINX Ingress Controller.

-----

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENCE.txt](https://www.google.com/search?q=service-mqtt/LICENCE.txt) file for details.