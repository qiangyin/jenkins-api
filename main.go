package main

import (
	"github.com/jenkins-api/api"
	"github.com/jenkins-api/g"
	"flag"
	"net/url"
	"fmt"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	g.ParseConfig(*cfg)
	jenkins := NewJenkinsWithData()
	jobs, err := jenkins.GetJobs(true)
	if err != nil {
	}

	if len(jobs) == 0 {
	}
	//	job, _ := jenkins.GetJob("abcd", true)
	//	jenkins.Build(job, nil, true)
	//	mavenJobItem := MavenJobItem()
	//	jenkins.CreateJob(mavenJobItem, "ab", true)
	//	jenkins.CreateJobByCfgXml("" +
	//		`<?xml version='1.0' encoding='UTF-8'?>
	//<project>
	//  <actions/>
	//  <description>测试jenkins git</description>
	//  <keepDependencies>false</keepDependencies>
	//  <properties>
	//    <com.dabsquared.gitlabjenkins.connection.GitLabConnectionProperty plugin="gitlab-plugin@1.4.3">
	//      <gitLabConnection></gitLabConnection>
	//    </com.dabsquared.gitlabjenkins.connection.GitLabConnectionProperty>
	//    <hudson.model.ParametersDefinitionProperty>
	//      <parameterDefinitions>
	//        <hudson.model.ChoiceParameterDefinition>
	//          <name>branch</name>
	//          <description></description>
	//          <choices class="java.util.Arrays$ArrayList">
	//            <a class="string-array">
	//              <string>dev </string>
	//              <string>master</string>
	//            </a>
	//          </choices>
	//        </hudson.model.ChoiceParameterDefinition>
	//        <hudson.model.ChoiceParameterDefinition>
	//          <name>env</name>
	//          <description></description>
	//          <choices class="java.util.Arrays$ArrayList">
	//            <a class="string-array">
	//              <string>300</string>
	//              <string>500</string>
	//              <string>800</string>
	//            </a>
	//          </choices>
	//        </hudson.model.ChoiceParameterDefinition>
	//        <hudson.model.StringParameterDefinition>
	//          <name>parameter</name>
	//          <description></description>
	//          <defaultValue>{&quot;default&quot;:0}</defaultValue>
	//        </hudson.model.StringParameterDefinition>
	//      </parameterDefinitions>
	//    </hudson.model.ParametersDefinitionProperty>
	//  </properties>
	//  <scm class="hudson.plugins.git.GitSCM" plugin="git@3.0.1">
	//    <configVersion>2</configVersion>
	//    <userRemoteConfigs>
	//      <hudson.plugins.git.UserRemoteConfig>
	//        <url>https://git.oschina.net/zedhs/sso-java.git</url>
	//        <credentialsId>43ea5e08-a999-4b8b-ac6a-17445f1d3b76</credentialsId>
	//      </hudson.plugins.git.UserRemoteConfig>
	//    </userRemoteConfigs>
	//    <branches>
	//      <hudson.plugins.git.BranchSpec>
	//        <name>*/master</name>
	//      </hudson.plugins.git.BranchSpec>
	//    </branches>
	//    <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
	//    <submoduleCfg class="list"/>
	//    <extensions/>
	//  </scm>
	//  <canRoam>true</canRoam>
	//  <disabled>false</disabled>
	//  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
	//  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
	//  <triggers/>
	//  <concurrentBuild>false</concurrentBuild>
	//  <builders/>
	//  <publishers/>
	//  <buildWrappers/>
	//</project>`, "abcd", true)
	//var build api.Build
	//build.Url = "http://10.128.31.109:8800/job/auto/10"
	//log, _ := jenkins.GetBuildConsoleOutput(build, true);
	//fmt.Println(string(log))
	job, _ := jenkins.GetJob("TestGit", true);
	//param := map[string][]string{
	//	"branch": []string{"dev"},
	//}
	//fmt.Println(param.en)
	a:=url.Values{}
	a.Set("branch","master")
	fmt.Println(a.Encode())
	jenkins.Build(job, a, true)

}
func NewJenkinsWithData() *api.Jenkins {
	var auth api.Auth
	auth.Username = "admin"
	auth.ApiToken = "admin"
	return api.NewJenkins(&auth, "http://10.128.31.109:8800")
}

func MavenJobItem() api.MavenJobItem {
	scm := api.Scm{
		ScmContent: api.ScmSvn{
			Locations: api.Locations{
				[]api.ScmSvnLocation{
					api.ScmSvnLocation{IgnoreExternalsOption: "false", DepthOption: "infinity", Local: ".", Remote: "http://some-svn-url"},
				},
			},
			IgnoreDirPropChanges: "false",
			FilterChangelog:      "false",
			WorkspaceUpdater:     api.WorkspaceUpdater{Class: "hudson.scm.subversion.UpdateUpdater"},
		},
		Class:  "hudson.scm.SubversionSCM",
		Plugin: "subversion@1.54",
	}
	triggers := api.Triggers{[]api.Trigger{api.ScmTrigger{}}}
	postStep := api.RunPostStepsIfResult{Name: "FAILURE", Ordinal: "2", Color: "RED", CompleteBuild: "true"}
	settings := api.JobSettings{Class: "jenkins.mvn.DefaultSettingsProvider"}
	globalSettings := api.JobSettings{Class: "jenkins.mvn.DefaultSettingsProvider"}
	jobItem := api.MavenJobItem{
		Plugin:               "maven-plugin@2.7.1",
		Description:          "test description",
		Scm:                  scm,
		Triggers:             triggers,
		RunPostStepsIfResult: postStep,
		Settings:             settings,
		GlobalSettings:       globalSettings,
	}
	return jobItem
}
