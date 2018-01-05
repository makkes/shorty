#!groovy

    properties([
            disableConcurrentBuilds()
    ])

    node {

        checkout scm

            docker.image('golang:1.8').inside {

                environment {
                    GIT_COMMITTER_EMAIL = 'jenkins@jenkins.makk.es'
                        GIT_COMMITTER_NAME = 'Jenkins'
                }

                stage('compile') {
                    sh 'go get && go build'
                }

                stage('build') {
                    sh 'docker build -t shorty:latest .'
                }
            }
    }
