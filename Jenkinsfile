#!groovy

    properties([
            disableConcurrentBuilds()
    ])

    node {

        checkout scm

            docker.image('golang:1.8').inside("-u root -v ${pwd()}:/go/src/github.com/makkes/shorty") {
                stage('compile') {
                    sh 'cd /go/src/github.com/makkes/shorty && go get && go build'
                }

                stage('build') {
                    sh 'cd /go/src/github.com/makkes/shorty && docker build -t shorty:latest .'
                }
            }
    }
