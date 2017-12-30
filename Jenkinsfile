#!groovy

properties([
        disableConcurrentBuilds()
])

node {

    docker.image('golang:1.8').inside {

        checkout scm

            stage('compile') {
                sh 'go build'
            }

        stage('build') {
            sh 'docker build -t shorty:latest .'
        }
    }
}
