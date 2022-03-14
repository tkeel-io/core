pipeline {
  agent {
    node {
      label 'base'
    }
  }
    parameters {
        string(name:'APP_VERSION',defaultValue: '0.4.2',description:'')
        string(name:'CHART_VERSION',defaultValue: '0.4.2',description:'')
    }

    environment {
        // Docker access token,定义在凭证中 
        DOCKER_CREDENTIAL_ID = 'dockerhub-tkeel'
        // GitHub access token,定义在凭证中
        GITHUB_CREDENTIAL_ID = 'github'
        // k8s kubeconfig,定义在凭证中
        KUBECONFIG_CREDENTIAL_ID = 'kubeconfig'
        // Docker 仓库
        REGISTRY = 'docker.io'
        // Docker 空间
        DOCKERHUB_NAMESPACE = 'tkeelio'
        // Github 账号
        GITHUB_ACCOUNT = 'tkeel-io'
        // 组件名称
        APP_NAME = 'core'
        // please ignore
        CHART_REPO_PATH = '/home/jenkins/agent/workspace/helm-charts'
    }

    stages {
        stage ('checkout scm') {
            steps {
                checkout(scm)
            }
        }
 
        stage ('build & push image') {
            steps {
                container ('base') {
                    sh 'docker build -f Dockerfile -t $REGISTRY/$DOCKERHUB_NAMESPACE/$APP_NAME:$BRANCH_NAME-$APP_VERSION .'
                    withCredentials([usernamePassword(passwordVariable : 'DOCKER_PASSWORD' ,usernameVariable : 'DOCKER_USERNAME' ,credentialsId : "$DOCKER_CREDENTIAL_ID" ,)]) {
                        sh 'echo "$DOCKER_PASSWORD" | docker login $REGISTRY -u "$DOCKER_USERNAME" --password-stdin'
                        sh 'docker push $REGISTRY/$DOCKERHUB_NAMESPACE/$APP_NAME:$BRANCH_NAME-$APP_VERSION'
                    }
                }
            }
        }

        stage('build & push chart'){
          steps {
              container ('base') {
                sh 'helm3 package charts/core --app-version=$APP_VERSION --version=$CHART_VERSION'
                // input(id: 'release-image-with-tag', message: 'release image with tag?')
                  withCredentials([usernamePassword(credentialsId: "$GITHUB_CREDENTIAL_ID", passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                    sh 'git config --global user.email "lunz1207@yunify.com"'
                    sh 'git config --global user.name "lunz1207"'
                    sh 'mkdir -p $CHART_REPO_PATH'
                    sh 'git clone https://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ACCOUNT/helm-charts.git $CHART_REPO_PATH'
                    sh 'mv ./$APP_NAME-*.tgz $CHART_REPO_PATH/$APP_NAME-$CHART_VERSION.tgz'
                    sh 'cd $CHART_REPO_PATH && helm3 repo index . --url=https://$GITHUB_ACCOUNT.github.io/helm-charts'
                    sh 'cd $CHART_REPO_PATH && git add . '
                    sh 'cd $CHART_REPO_PATH && git commit -m "feat:update chart"'
                    sh 'cd $CHART_REPO_PATH && git push https://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ACCOUNT/helm-charts.git'
                  }
              }
          }
        }
    }
}