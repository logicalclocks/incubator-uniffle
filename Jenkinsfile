node("local") {
  def dockerRegistry = 'n59k7749.c1.de1.container-registry.ovh.net'
  def version = readFile("${env.WORKSPACE}/version.txt").trim()
  def controllerImage = "${dockerRegistry}/hopsworks/rss-controller:${version.trim()}"
  def webhookImage = "${dockerRegistry}/hopsworks/rss-webhook:${version.trim()}"
  def uniffleVersion = "0.10.2"

  stage('Clone repository') {
      checkout scm
  }

  stage('Build and push images to registry') {

    withCredentials([usernamePassword(credentialsId: 'a0770738-4ef3-4acc-a6ba-097ee6c85b44', passwordVariable: 'PASSWORD', usernameVariable: 'USERNAME')]) {
      sh """
        set -ex
        git status
        docker login -u ${USERNAME} -p ${PASSWORD} $dockerRegistry
        mkdir -p "\$WORKSPACE@tmp"
        cat > "\$WORKSPACE@tmp/mvn-settings.xml" <<EOF
<settings>
  <servers>
    <server>
      <id>HopsEE</id>
      <username>${USERNAME}</username>
      <password>${PASSWORD}</password>
    </server>
  </servers>
</settings>
EOF

        docker run --rm \
          -v .:/incubator-uniffle \
          -v "\$WORKSPACE@tmp/mvn-settings.xml:/tmp/mvn-settings.xml:ro" \
          -w /incubator-uniffle \
          eclipse-temurin:17-jdk \
          /bin/bash build_distribution.sh --spark4-profile spark4.1 --hadoop-profile hadoop3.2 --without-mr --without-tez --without-spark2 --without-spark3 -s /tmp/mvn-settings.xml -U

        cd deploy/kubernetes/docker ||  exit
        ./build.sh --hadoop-version 3.2.0.15-EE-SNAPSHOT --registry $dockerRegistry --nexus-user $USERNAME --nexus-password $PASSWORD --push-image true
        cd ../../..

        mkdir -p /opt/repository/master/rss/$version/
        cp  client-spark/spark3-shaded/target/rss-client-spark3-shaded-${uniffleVersion}.jar /opt/repository/master/rss/${version}/rss-client-spark4-shaded-${version}.jar

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
