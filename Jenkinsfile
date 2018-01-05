#!groovy

    properties([
            disableConcurrentBuilds()
    ])

    node {

        checkout scm

            docker.image('golang:1.8').withRun('-u root') {
                stage('compile') {
                    sh 'go get && go build'
                }

                stage('build') {
                    sh 'docker build -t shorty:latest .'
                }
            }
    }
