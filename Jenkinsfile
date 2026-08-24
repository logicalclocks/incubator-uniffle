node("local") {
  def dockerRegistry = 'n59k7749.c1.de1.container-registry.ovh.net'
  def uniffleVersion = "0.10.1"

  stage('Clone repository') {
      checkout scm
  }

  def version = readFile("${env.WORKSPACE}/version.txt").trim()
  def controllerImage = "${dockerRegistry}/hopsworks/rss-controller:${version.trim()}"
  def webhookImage = "${dockerRegistry}/hopsworks/rss-webhook:${version.trim()}"

  stage('Build and push images to registry') {

    withCredentials([usernamePassword(credentialsId: 'a0770738-4ef3-4acc-a6ba-097ee6c85b44', passwordVariable: 'PASSWORD', usernameVariable: 'USERNAME')]) {
      sh """
        set -ex
        git status
        docker login -u ${USERNAME} -p ${PASSWORD} $dockerRegistry

        docker run --rm -v .:/incubator-uniffle -w /incubator-uniffle  eclipse-temurin:8-jdk /bin/bash build_distribution.sh --spark3-profile spark3.5 --hadoop-profile hadoop3.2 --without-mr --without-tez --without-spark2

        cd deploy/kubernetes/docker ||  exit
        ./build.sh --hadoop-version 3.4.3.2-EE-RC2 --hadoop-profile hadoop3.2 --registry $dockerRegistry --nexus-user $USERNAME --nexus-password $PASSWORD --push-image true
        cd ../../..

        mkdir -p /opt/repository/master/rss/$version/
        cp  client-spark/spark3-shaded/target/rss-client-spark3-shaded-${uniffleVersion}.jar /opt/repository/master/rss/${version}/rss-client-spark3-shaded-${version}.jar

        # build the controller and webhook images
        cd deploy/kubernetes/operator ||  exit 1
        docker build . --progress=plain -t $controllerImage --build-arg MODULE=controller -f hack/Dockerfile
        docker build . --progress=plain -t $webhookImage --build-arg MODULE=webhook -f hack/Dockerfile
        # push the controller and webhook images
        docker push $controllerImage
        docker push $webhookImage
      """
    }
  }
}
