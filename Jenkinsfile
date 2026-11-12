@Library("jenkins-library@main")

import com.logicalclocks.jenkins.k8s.ECR
import com.logicalclocks.jenkins.k8s.ImageBuilder

// Builds the rss distribution, publishes the shaded client jar to repo.hops.works and
// builds the hopsworks/rss, hopsworks/rss-controller and hopsworks/rss-webhook images.
//
// The images go through the shared ImageBuilder, like the hive and docker-images jobs:
// it builds and pushes each image to every registry it is configured with (the private
// registry and the public ECR dev and prod registries) and creates a missing ECR
// repository. The tag is version.txt.
//
// Before anything is built the job looks up every image of that version in every
// registry. All present: the job stops (SKIP), so the jar under
// /opt/repository/master/rss/<version>/ and the images always come from the same commit.
// None present: a normal build. Some present, after a run that failed half way: the
// build runs, ImageBuilder's immutability check leaves the published images alone and
// pushes the missing ones, and the published jar is kept too. Bump version.txt for a
// change that should ship, or run with FORCE_BUILD to republish the current version
// everywhere.
//
// Written with plain loops and String methods on purpose: the Jenkins script sandbox
// does not approve the Groovy collection helpers (collect, findAll, find, plus on a
// list) by default, and a build would stop on the first use until an administrator
// approves the signature. Same rule as hopsworks-ee/docker/Jenkinsfile.
properties([
  parameters([
    booleanParam(
      name: 'FORCE_BUILD',
      defaultValue: false,
      description: 'Build and publish even if every image of version.txt is already in every registry. The jar on repo.hops.works and the images in every registry are overwritten with this commit; a registry that refuses to overwrite a tag still wins.'
    )
  ])
])

// Every <image>:<version> that is missing from a registry ImageBuilder pushes to, as
// the full reference. Authenticates against each registry on the way, as ImageBuilder
// does before it pushes.
def missingImages(builder, List<String> images, String version) {
  def missing = []
  for (def r : builder.REGISTRIES) {
    r.auth()
    for (int i = 0; i < images.size(); i++) {
      if (r instanceof ECR) {
        // imageExists only recognises a missing tag; a missing repository is an error.
        r.ensureRepository(images[i])
      }
      def ref = r.buildImageName(images[i], version)
      if (!builder.imageExists(ref)) {
        missing.add(ref)
      }
    }
  }
  return missing
}

// One item per line, indented, for an echo.
def asLines(List<String> items) {
  def text = ""
  for (int i = 0; i < items.size(); i++) {
    text += "\n  " + items[i]
  }
  return text
}

// image-build.env, written by build.sh --prepare-only: one KEY=VALUE per line, the
// build arguments of the rss image. Returned as the list withEnv takes.
def readImageBuildEnv(String path) {
  def envList = []
  def lines = readFile(path).split("\n")
  for (int i = 0; i < lines.length; i++) {
    def line = lines[i].trim()
    if (line && !line.startsWith('#')) {
      envList.add(line)
    }
  }
  return envList
}

node("local") {
  def uniffleVersion = "0.10.1"
  // The Hadoop distribution shipped in the rss image.
  def hadoopVersion = "3.2.0.15-EE-SNAPSHOT"
  // The images of the two build manifests, by name.
  def images = ["hopsworks/rss", "hopsworks/rss-controller", "hopsworks/rss-webhook"]

  stage('Clone repository') {
      checkout scm
  }

  def version = readFile("${env.WORKSPACE}/version.txt").trim()
  def builder = new ImageBuilder(this)
  def publish = true
  def overwriteJars = true

  withCredentials([usernamePassword(credentialsId: 'a0770738-4ef3-4acc-a6ba-097ee6c85b44', passwordVariable: 'PASSWORD', usernameVariable: 'USERNAME')]) {
    stage('Check published version') {
      // FORCE_BUILD rewrites the current version everywhere. With the check left on,
      // ImageBuilder would skip every registry that already holds the tag, so the jar
      // would move and the images would not.
      builder.immutable = !params.FORCE_BUILD
      // auth() and imageExists() called from the script need the builder's isolated
      // Docker config (jenkins-library README, "Docker authentication scope").
      def missing = builder.withDockerConfig { missingImages(builder, images, version) }
      def total = images.size() * builder.REGISTRIES.size()
      if (params.FORCE_BUILD) {
        echo "FORCE  ${version} is rebuilt and pushed to every registry; ${missing.size()} of ${total} images are missing today."
      } else if (missing.size() == 0) {
        publish = false
        echo "SKIP   every image of ${version} is in every registry. Bump version.txt to ship this commit, or run with FORCE_BUILD."
      } else if (missing.size() == total) {
        echo "BUILD  ${version} is not published yet."
      } else {
        // The published images and jar came from an earlier run of this version and
        // stay as they are; only the missing images are pushed.
        overwriteJars = false
        echo "RESUME ${missing.size()} of ${total} images of ${version} are missing and will be pushed:" + asLines(missing)
      }
    }

    if (publish) {
      stage('Build distribution') {
        sh """
          set -ex
          git status

          docker run --rm -v .:/incubator-uniffle -w /incubator-uniffle  eclipse-temurin:8-jdk /bin/bash build_distribution.sh --spark3-profile spark3.5 --hadoop-profile hadoop3.2 --without-mr --without-tez --without-spark2

          # publish_jar <built jar> <published jar>: a jar that is already published is
          # kept on a resumed run, it belongs with the images that were pushed before.
          publish_jar() {
            if [ -f "\$2" ] && [ "${overwriteJars}" != "true" ]; then
              echo "keeping published \$2"
            else
              cp "\$1" "\$2"
            fi
          }
          mkdir -p /opt/repository/master/rss/${version}/
          publish_jar client-spark/spark3-shaded/target/rss-client-spark3-shaded-${uniffleVersion}.jar /opt/repository/master/rss/${version}/rss-client-spark3-shaded-${version}.jar
        """
      }

      stage('Prepare rss image context') {
        // Downloads the Hadoop distribution, stages the rss tarball next to the Dockerfile
        // and writes the build arguments to image-build.env. The image itself is built below.
        dir('deploy/kubernetes/docker') {
          sh """
            set -ex
            ./build.sh --hadoop-version ${hadoopVersion} --nexus-user \$USERNAME --nexus-password \$PASSWORD --prepare-only true
            cat image-build.env
          """
        }
      }

      stage('Build and push images') {
        // The rss manifest lists the image-build.env keys as its `variables`, which fill
        // the ${VAR} placeholders of its extraDockerArgs; the --build-arg flags themselves
        // are written in extraDockerArgs.
        def imageBuildEnv = readImageBuildEnv("${env.WORKSPACE}/deploy/kubernetes/docker/image-build.env")
        imageBuildEnv.add("RSS_IMAGE_VERSION=${version}")

        withEnv(imageBuildEnv) {
          dir('deploy/kubernetes/docker') {
            builder.run(readFile('build-manifest.json'))
          }
          // The operator Dockerfile lives in hack/ but builds from the module root, so the
          // manifest keeps the build in this directory (changeDirectory false).
          dir('deploy/kubernetes/operator') {
            builder.run(readFile('build-manifest.json'))
          }
        }
      }
    }
  }
}
