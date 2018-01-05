#!groovy

properties([
        disableConcurrentBuilds()
])

node {

    checkout scm

    docker.image('golang:1.8').inside {

            stage('compile') {
                sh 'go get && go build'
            }

        stage('build') {
            sh 'docker build -t shorty:latest .'
        }
    }
}
