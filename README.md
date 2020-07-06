# Better CI

Just a Better CI that you always wanted. Allows you to deploy PRs temporarily for a preview. Works with Docker-Compose and K8s deployments.

## Start CI
1. Cone Repo
```sh
$ git clone https://github.com/gdsoumya/better_ci.git
$ cd better_ci
```
2. Create .env file with : [**All fields are compulsory**]
```
PORT=<port-to-run-ci-server>
ACCESS_KEY=<your-personal-github-access-key>
DOCKER_USER=<docker-username-for-builds>
DOCKER_PASS=<docker-password>
WEBHOOK_SECRET=<webhook-secret>
```
3. Start CI Server
    ```sh
    $ ./output/better_ci
    ```
    or
    ```sh
    $ go run . 
    ```

## Setup Repo

1. Add a Web Hook to your repo with the CI server URL

```
WebHook URL : <server_url>:8080/webhook
```

2. Add Better CI Config to deploy preview builds

    Add a `.betterci/` dir at the root of your repo with a `config.json` file.
    
    Example `config.json`
    
* Docker-Compose Version : `.betterci/config.json`
```json
{
    "cmd":["echo hello","ls -al"],
    "build":[{
        "name":"nginx",
        "file":"app/frontend/Dockerfile.client-local",
        "context":"app/frontend",
        "push": false
        },{
        "name":"server",
        "file":"app/backend/build/Dockerfile.server",
        "context":"app/backend",
        "push":true
    }],
    "docker-compose":"app/better-ci-compose.yml"
}
```   

* K8s Version : `.betterci/config.json`
```json
{
    "cmd":["echo hello","ls -al"],
    "build":[{
        "name":"nginx",
        "file":"app/frontend/Dockerfile.client-local",
        "context":"app/frontend",
        "push": true
        },{
        "name":"server",
        "file":"app/backend/build/Dockerfile.server",
        "context":"app/backend",
        "push":true
    }],
    "k8s-manifest":"litmus-portal/k8s-manifest.yml"
}
``` 

### Config File Doc
1. **cmd** : A string array containing simple bash commands to execute, could be used to run tests etc. Currently compound commands like `cd dir && ls` are not allowed and also `cd` command isn't possible to execute(limitation of go:os/exec)
2. **build** : An object array, containing details of the images to be built dynamically during preview deployment.<br>The Object contains the following fields:
```
{
    "name" : "<name for the image>",
    "file" : "<relative file path to Dockerfile template>",
    "context" : "<relative path of context for build>",
    "push" : <bool to push to docker hub, used for k8s-manifests>
}
```
The field `name` can be later reused in the Docker-Compose or K8s-Manifest Templates as variables to use the dynamically generated images for the deployment.

**All relative paths are with respect to the root of the repository.**
    
3. `docker-compose` or `k8s-manifest` : A string denoting the relative path to the docker-compose file template or k8s-maifest template, both cannot be used at the same time. If used either error will be thrown or docker-compose will be given preference.

### Docker-Compose Template
The Docker compose template allows 2 variables :
1. **#{image-name}** : `image-name` is the name of the image you mentioned in the `build` stage, which is dynamically generated during build. It can be used to build the current image of the PR and use it, without pushing or storing it permanently.
2. **#{PORT}** : `PORT` can be used to dynamically select the port to expose a container service, it is adviced to use this when you have to expose port on the system to access the container services. The `PORT` value will be added to the comment in the PR after preview deployment, for access.

* **Remember not to give any container-name, it will be auto generated**

Example Docker-Compose Template :
```
version: '3'

services:
    nginx-server:
        ports:
            - #{PORT}:80
	image: #{nginx}
    backend-server:
	image: #{server}
```  
### K8s Template
The K8s deployments only works with `nodePort` Services and allows the following variables:
1. **#{image-name}** : `image-name` is the name of the image you mentioned in the `build` stage, which is dynamically generated during build. It can be used to build the current image of the PR and use it, without pushing or storing it permanently.

* **Remember to use `"push":true` in the `build` stage for the images used in the k8s deployment**
* **The namespaces for the deployments are auto-generated hence,specifying namespace can cause errors**
* **Do not mention the `nodePort` value in the template, as it may cause conflicts with other deployments**

Example K8s Template :
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-server
  labels:
    component: backend-server
spec:
  replicas: 1
  selector:
    matchLabels:
      component: backend-server
  template:
    metadata:
      labels:
        component: backend-server
    spec:
      containers:
      - name: backend-server
        image: #{server}
        ports:
        - containerPort: 8080
        imagePullPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: backend-server-service
spec:
  type: NodePort
  ports:
    - name: http
      port: 9002
      targetPort: 8080
  selector:
    component: backend-server
```

**The templates can be separately defined from the actual deployment files, so you can have a `docker-compose.yml` and a `ci-docker-compose.yml` at the same time. Just make sure to mention the correct file path in the -`.betterci/config.json`**

## Initiate Build

To start the preview build of a PR in a registered repo -
1. Comment In the PR: **/preview**

This will run the ci with the specified config and deploy a preview that will expire in 5mins. The links of the exposed services in both docker-compose and k8s case will be added to the comment automatically once they are deployed. Once the link expires an expired message gets edited into the comment itself.

Currently, only one build can be active per PR, so once a build is initiated you have to wait before it expires to perform a new build of the PR. 

## Author
* Soumya Ghosh Dastidar
