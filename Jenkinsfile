   agent any       
				stage('Test on Linux') {
            agent {   //只用于运行该stage的agent
                label 'linux'
            }
            steps {
                unstash 'app' 
                sh 'make check'
            }
            post { //stage内的post action
                always {
                    junit '**/target/*.xml'
                }
            }
        }