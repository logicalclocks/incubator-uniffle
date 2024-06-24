pipeline {
    agent {
        label "local"
    }
    environment {
        VERSION = readFile "${env.WORKSPACE}/version.txt"
        BUILD_BRANCH = readFile "${env.WORKSPACE}/build_branch.txt"
    }
    stages {
        stage("checkout") {
            steps {
                sh """
                    set -ex
                    git fetch --all
                    git checkout ${BUILD_BRANCH}
                    git reset --hard origin/${BUILD_BRANCH}
                """
            }
        }
        stage("build and publish") {
            agent {
                label "local"
            }
            steps {
             withCredentials([usernamePassword(credentialsId: 'cred', passwordVariable: 'NEXUS_CREDS_PSW', usernameVariable: 'NEXUS_CREDS_USR')]) {
                sh """
                    set -ex
                    echo "Building RSS version ${VERSION} on branch ${BUILD_BRANCH}"
                    docker login -u ${NEXUS_CREDS_USR} -p ${NEXUS_CREDS_PSW} docker.hops.works

                    ./build_distribution.sh --spark3-profile spark3 --hadoop-profile hadoop3.2 --without-dashboard
                    cd deploy/kubernetes/docker ||  exit
                    ./build.sh --hadoop-version 3.2.0.13-EE-SNAPSHOT --registry docker.hops.works --nexus-user $NEXUS_CREDS_USR --nexus-password $NEXUS_CREDS_PSW
                    cd ../../..

                    mkdir -p /opt/repository/master/rss/${VERSION}/
                    cp  client-spark/spark3-shaded/target/rss-client-spark3-shaded-${VERSION}.jar /opt/repository/master/rss/${VERSION}/
                """
             }
            }
        }
    }
}