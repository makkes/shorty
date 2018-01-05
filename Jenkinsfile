#!groovy

properties([
        disableConcurrentBuilds()
])

node {

    docker.image('golang:1.8').inside {

        checkout scm

            stage('compile') {
                sh 'git config --global user.name "Jenkins" && git config --global user.email "jenkins@jenkins.makk.es"''
                sh 'go get && go build'
            }

        stage('build') {
            sh 'docker build -t shorty:latest .'
        }
    }
}
