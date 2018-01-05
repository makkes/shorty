#!groovy

    properties([
            disableConcurrentBuilds()
    ])

    node {

        stage('checkout code') {
            checkout scm
        }

        stage('compile') {
            sh 'pwd'
            sh 'ls -lh'
            docker.image('golang:1.8').inside("-u root -v ${pwd()}:/go/src/github.com/makkes/shorty") {
                sh 'cd /go/src/github.com/makkes/shorty && ls -lh'
                sh 'cd /go/src/github.com/makkes/shorty && go get && go build'
            }
        }

        stage('build') {
            sh 'docker build -t shorty:latest .'
        }
    }
