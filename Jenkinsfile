    def exitNow = false
    def skipBuild = true

    def imagename = "chainsource-gateway"
    def version = "1.0.0-$BUILD_ID"
    def buildSlave = (params.BUILDSLAVE != null) ? params.BUILDSLAVE : 'Blockchain1'
    def namespace = (params.NAMESPACE != null) ? params.NAMESPACE : 'dbom'
    def branch = (params.BRANCH != null) ? params.BRANCH : 'master'
    def additionalHelmParams = ""
	def imagePullSecretName = "robot-rancher"

    if(params.ENABLE_JAEGER){
    echo "Parameter selected for injecting jaeger sidecar"
    additionalHelmParams = "--set jaeger.enabled=true --set jaeger.agent.sidecar.enabled=true --set jaeger.agent.sidecar.name=jaeger-blockchain"
    }

    node("${BuildSlave}") {
        // Get the code from Bitbucket repository
        stage('Cloning from Git')  {
            git branch: "${branch}", changelog: true, credentialsId: '3997416c-05b5-438a-aee1-0f382e47479c', url: 'https://ustr-bitbucket-1.na.uis.unisys.com:8443/scm/et/prod-dbom.git'
            skipBuild = sh (script: "git log -1 --pretty=format:%s#%b | grep '\\[skip ci\\]'", returnStatus: true) == 0
            if (skipBuild) {
                currentBuild.result = 'ABORTED'
                exitNow = true
            }
        }

        // Exit pipeline if so directed
        if (exitNow) {
            return
        }
        dir("chainsource-gateway"){
            stage('Docker Build, Test and Push') {
                sh "docker build --tag ustr-harbor-1.na.uis.unisys.com/blockchain/${imagename}:${version} ."
                sh "docker push ustr-harbor-1.na.uis.unisys.com/blockchain/${imagename}:${version}"
            }

            stage('Helm Build and Install') {
                try {
                    sh "helm del ${imagename} -n ${namespace}"
                }
                catch (err) {
                    echo "release ${imagename} does not exist"
                }
                sh "sleep 10s"
                sh "helm install ${imagename} ./chainsource-gateway --wait --version ${version} --namespace ${namespace} --set image.tag=${version},namespace=${namespace},imagePullSecrets[0].name=${imagePullSecretName} ${additionalHelmParams}"
            }

            stage('Clean up workspace') {
                echo 'Deleting Workspace'
                cleanWs deleteDirs: true
           }
        }
    }
