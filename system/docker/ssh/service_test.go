package docker_test

import (
	"fmt"
	"github.com/viant/endly/system/docker/docker"
	"log"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/assertly"
	"github.com/viant/endly"
	"github.com/viant/endly/system/exec"
	"github.com/viant/endly/util"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"github.com/viant/toolbox/url"
)

func TestDockerService_ComposeUp(t *testing.T) {

	var credentialFile, err = util.GetDummyCredential()
	assert.Nil(t, err)

	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	manager := endly.New()

	useCases := []struct {
		description string
		baseDir     string
		target      *url.Resource
		request     *ssh.ComposeRequestUp
		expected    interface{}
		HasError    bool
	}{

		{
			description: "Docker Compose up request",
			target:      target,
			baseDir:     "test/compose/up/darwin",
			request:     &ssh.ComposeRequestUp{&ssh.ComposeRequest{Target: target, Source: url.NewResource("test/compose/up/docker-compose.yaml")}, true},
			expected: `{
  "Containers": [
    {	
      "ContainerID":"5280cb455e33",
      "Image": "redis",
      "Command": "docker-entrypoint.s…",
      "Status": "up",
      "Port": "6379/tcp",
      "Names": "redis"
    }
  ]
}
`,
		},
	}

	for _, useCase := range useCases {
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		if !assert.Nil(t, err) {
			log.Fatal(err)
		}

		var response = &ssh.ComposeResponse{}
		err = endly.Run(context, useCase.request, response)
		if useCase.HasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			t.Log(err.Error())
			continue
		}
		assertly.AssertValues(t, useCase.expected, response, useCase.description)
	}
}

func TestDockerService_ComposeDown(t *testing.T) {

	var credentialFile, err = util.GetDummyCredential()
	assert.Nil(t, err)

	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	manager := endly.New()

	useCases := []struct {
		description string
		baseDir     string
		target      *url.Resource
		request     *ssh.ComposeRequestDown
		expected    interface{}
		HasError    bool
	}{

		{
			description: "Docker Compose down request",
			target:      target,
			baseDir:     "test/compose/down/darwin",
			request:     &ssh.ComposeRequestDown{&ssh.ComposeRequest{Target: target, Source: url.NewResource("test/compose/up/docker-compose.yaml")}},
			expected:    `{"Containers":[]}`,
		},
	}

	for _, useCase := range useCases {
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		defer context.Close()
		if !assert.Nil(t, err) {
			log.Fatal(err)
		}

		var response = &ssh.ComposeResponse{}
		err = endly.Run(context, useCase.request, response)
		if useCase.HasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			t.Log(err.Error())
			continue
		}
		assertly.AssertValues(t, useCase.expected, response, useCase.description)
	}

}

func TestDockerService_Images(t *testing.T) {
	var credentialFile, err = util.GetDummyCredential()
	assert.Nil(t, err)

	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	var manager = endly.New()
	var useCases = []struct {
		baseDir    string
		target     *url.Resource
		Repository string
		Tag        string
		Expected   []*ssh.ImageInfo
	}{
		{
			"test/images/darwin",
			target,
			"mysql",
			"5.6",
			[]*ssh.ImageInfo{
				{
					Repository: "mysql",
					Tag:        "5.6",
					ImageID:    "96dc914914f5",
					Size:       313524224,
				},
			},
		},
		{
			"test/images/darwin",
			target,
			"",
			"",
			[]*ssh.ImageInfo{
				{
					Repository: "mysql",
					Tag:        "5.6",
					ImageID:    "96dc914914f5",
					Size:       313524224,
				},
				{
					Repository: "mysql",
					Tag:        "5.7",
					ImageID:    "5709795eeffa",
					Size:       427819008,
				},
			},
		},
		{
			"test/images/linux",
			target,
			"mysql",
			"5.6",
			[]*ssh.ImageInfo{
				{
					Repository: "mysql",
					Tag:        "5.6",
					ImageID:    "96dc914914f5",
					Size:       313524224,
				},
			},
		},
	}

	for _, useCase := range useCases {
		context, err := exec.NewSSHReplayContext(manager, useCase.target, useCase.baseDir)
		if assert.Nil(t, err) {

			if assert.Nil(t, err) {

				var request = ssh.NewImagesRequest(useCase.target, useCase.Repository, useCase.Tag)
				var response = &ssh.ImagesResponse{}
				err := endly.Run(context, request, response)
				var baseCase = useCase.baseDir + " " + useCase.Repository
				if !assert.Nil(t, err, baseCase) {
					return
				}
				if len(response.Images) != len(useCase.Expected) {
					assert.Fail(t, fmt.Sprintf("Expected %v image info but had %v", len(useCase.Expected), len(response.Images)), useCase.baseDir)
				}
				for i, expected := range useCase.Expected {
					if i >= len(response.Images) {
						assert.Fail(t, fmt.Sprintf("Image info was missing [%v] %v", i, baseCase))
						continue
					}
					var actual = response.Images[i]
					assert.Equal(t, expected.Tag, actual.Tag, "Tag "+baseCase)
					assert.EqualValues(t, expected.ImageID, actual.ImageID, "ImageID "+baseCase)
					assert.Equal(t, expected.Repository, actual.Repository, "Repository "+baseCase)
					assert.EqualValues(t, expected.Size, actual.Size, "Size "+baseCase)

				}
			}

		}

	}
}

func TestDockerService_Run(t *testing.T) {

	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)

	mySQLcredentialFile, err := util.GetCredential("mysql", "root", "dev")
	assert.Nil(t, err)

	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	var manager = endly.New()
	var useCases = []struct {
		baseDir    string
		Request    *ssh.RunRequest
		Expected   *ssh.ContainerInfo
		TargetName string
		Error      string
	}{
		{
			"test/run/existing/darwin",
			&ssh.RunRequest{
				Target: target,
				Image:  "mysql:5.6",
				Ports: map[string]string{
					"3306": "3306",
				},
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": "**mysql**",
				},
				Mount: map[string]string{
					"/tmp/my.cnf": "/etc/my.cnf",
				},
				Secrets: map[string]string{
					"**mysql**": mySQLcredentialFile,
				},
			},
			&ssh.ContainerInfo{
				Status:      "up",
				Names:       "testMysql",
				ContainerID: "83ed7b545cbf",
			},
			"testMysql",
			"",
		},
		{
			"test/run/new/darwin",
			&ssh.RunRequest{

				Target: target,
				Image:  "mysql:5.6",
				Ports: map[string]string{
					"3306": "3306",
				},
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": "**mysql**",
				},
				Mount: map[string]string{
					"/tmp/my.cnf": "/etc/my.cnf",
				},
				Secrets: map[string]string{
					"**mysql**": mySQLcredentialFile,
				},
			},
			&ssh.ContainerInfo{
				Status:      "up",
				Names:       "testMysql",
				ContainerID: "98a28566ba7a",
			},
			"testMysql",
			"",
		},
		{
			"test/run/error/darwin",
			&ssh.RunRequest{
				Target: target,
				Image:  "mysql:5.6",
				Ports: map[string]string{
					"3306": "3306",
				},
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": "**mysql**",
				},
				Mount: map[string]string{
					"/tmp/my.cnf": "/etc/my.cnf",
				},
				Secrets: map[string]string{
					"**mysql**": mySQLcredentialFile,
				},
			},
			&ssh.ContainerInfo{},
			"testMysql01",
			"error executing docker run --name testMysql01 -e MYSQL_ROOT_PASSWORD=**mysql** -v /tmp/my.cnf:/etc/my.cnf -p 3306:3306  -d mysql:5.6 , c3d9749a1dc43332bb5a58330187719d14c9c23cee55f583cb83bbb3bbb98a80\ndocker: Error response from daemon: driver failed programming external connectivity on endpoint testMysql01 (5c9925d698dfee79f14483fbc42a3837abfb482e30c70e53d830d3d9cfd6f0da): Error starting userland proxy: Bind for 0.0.0.0:3306 failed: port is already allocated.\n at docker.run",
		},
		{
			"test/run/active/darwin",
			ssh.NewRunRequest(target, "testMysql",
				map[string]string{
					"**mysql**": mySQLcredentialFile,
				}, "mysql:5.6", "", map[string]string{
					"MYSQL_ROOT_PASSWORD": "**mysql**",
				}, map[string]string{
					"/tmp/my.cnf": "/etc/my.cnf",
				}, map[string]string{
					"3306": "3306",
				}, nil, ""),
			&ssh.ContainerInfo{
				Status:      "up",
				Names:       "testMysql",
				ContainerID: "84df38a810f7",
			},
			"testMysql",
			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		if assert.Nil(t, err) {

			defer context.Close()
			if assert.Nil(t, err) {
				useCase.Request.Name = useCase.TargetName
				var response = &ssh.RunResponse{}
				err := endly.Run(context, useCase.Request, response)
				var description = useCase.baseDir + " " + useCase.TargetName
				if useCase.Error != "" {
					assert.EqualValues(t, useCase.Error, fmt.Sprintf("%v", err), description)
					continue
				}
				if !assert.Nil(t, err) {
					continue
				}
				var expected = useCase.Expected
				assert.EqualValues(t, expected.Status, response.Status, "Status "+description)
				assert.EqualValues(t, expected.Names, response.Names, "Names "+description)
				assert.EqualValues(t, expected.ContainerID, response.ContainerID, "ContainerID "+description)
			}

		}

	}
}

func TestDockerService_ExecRequest(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)

	mySQLcredentialFile, err := util.GetCredential("mysql", "root", "dev")
	assert.Nil(t, err)

	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	var manager = endly.New()
	var useCases = []struct {
		baseDir    string
		Request    *ssh.ExecRequest
		Expected   string
		TargetName string
		Error      string
	}{
		{
			"test/command/export/darwin",
			ssh.NewExecRequest(ssh.NewBaseRequest(target, "testMysql"),
				"mysqldump  -uroot -p***mysql*** --all-databases --routines | grep -v 'Warning' > /tmp/dump.sql",
				map[string]string{
					"***mysql***": mySQLcredentialFile,
				}, true, true, false),
			"",
			"testMysql",
			"",
		},
		{
			"test/command/import/darwin",
			&ssh.ExecRequest{
				BaseRequest: &ssh.BaseRequest{
					Target: target,
				},
				Interactive: true,
				Secrets: map[string]string{
					"**mysql**": mySQLcredentialFile,
				},
				Command: "mysql  -uroot -p**mysql** < /tmp/dump.sql",
			},
			"\r\nWarning: Using a password on the command line interface can be insecure.",
			"testMysql",
			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		if assert.Nil(t, err) {

			if assert.Nil(t, err) {
				useCase.Request.Name = useCase.TargetName
				var description = useCase.baseDir + " " + useCase.TargetName
				response := &ssh.ExecResponse{}
				err := endly.Run(context, useCase.Request, response)
				if useCase.Error != "" {
					assert.EqualValues(t, useCase.Error, fmt.Sprintf("%v", err), description)
					continue
				}
				if !assert.Nil(t, err) {
					continue
				}
				var expected = useCase.Expected
				assert.EqualValues(t, expected, response.Stdout, "Status "+description)
			}
		}
	}
}

func TestDockerService_Pull(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	var manager = endly.New()
	var useCases = []struct {
		baseDir  string
		Request  *ssh.PullRequest
		Expected *ssh.ImageInfo
		Error    string
	}{
		{
			"test/pull/linux",
			&ssh.PullRequest{
				Target:     target,
				Repository: "mysql",
				Tag:        "5.7",
			},
			&ssh.ImageInfo{
				Repository: "mysql",
				Tag:        "5.7",
				ImageID:    "5709795eeffa",
				Size:       427819008,
			},

			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		if assert.Nil(t, err) {
			service, err := context.Service(ssh.ServiceID)
			assert.Nil(t, err)
			if assert.Nil(t, err) {
				serviceResponse := service.Run(context, useCase.Request)

				var baseCase = useCase.baseDir + " "
				assert.Equal(t, useCase.Error, serviceResponse.Error, baseCase)

				actual, ok := serviceResponse.Response.(*ssh.PullResponse)
				if !ok {
					assert.Fail(t, fmt.Sprintf("process serviceResponse was empty %v %T", baseCase, serviceResponse.Response))
					continue
				}

				if actual == nil {
					continue
				}
				var expected = useCase.Expected
				assert.Equal(t, expected.Tag, actual.Tag, "Tag "+baseCase)
				assert.EqualValues(t, expected.ImageID, actual.ImageID, "ImageID "+baseCase)
				assert.Equal(t, expected.Repository, actual.Repository, "Repository "+baseCase)
				assert.EqualValues(t, expected.Size, actual.Size, "Size "+baseCase)

			}

		}

	}

}

func TestDockerService_Status(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //

	var manager = endly.New()
	var useCases = []struct {
		baseDir  string
		Request  *ssh.ContainerStatusRequest
		Expected *ssh.ContainerStatusResponse
		Error    string
	}{
		{
			"test/status/up/linux",
			&ssh.ContainerStatusRequest{
				Target: target,
			},
			&ssh.ContainerStatusResponse{
				Containers: []*ssh.ContainerInfo{
					{
						ContainerID: "b5bcc949f075",
						Port:        "0.0.0.0:3306->3306/tcp",
						Command:     "docker-entrypoint...",
						Image:       "mysql:5.6",
						Status:      "up",
						Names:       "db1",
					},
				},
			},

			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		exec.GetReplayService(useCase.baseDir)
		if assert.Nil(t, err) {
			service, err := context.Service(ssh.ServiceID)
			assert.Nil(t, err)
			defer context.Close()
			if assert.Nil(t, err) {
				serviceResponse := service.Run(context, useCase.Request)

				var baseCase = useCase.baseDir + " "
				assert.Equal(t, useCase.Error, serviceResponse.Error, baseCase)

				response, ok := serviceResponse.Response.(*ssh.ContainerStatusResponse)
				if !ok {
					assert.Fail(t, fmt.Sprintf("process resonse was empty %v %T", baseCase, serviceResponse.Response))
					continue
				}

				var expected = useCase.Expected.Containers[0]
				var actual = response.Containers[0]

				assert.Equal(t, expected.ContainerID, actual.ContainerID, "ContainerID "+baseCase)
				assert.EqualValues(t, expected.Port, actual.Port, "Port "+baseCase)
				assert.Equal(t, expected.Command, actual.Command, "Command "+baseCase)
				assert.EqualValues(t, expected.Image, actual.Image, "Image "+baseCase)
				assert.EqualValues(t, expected.Names, actual.Names, "Names "+baseCase)
				assert.EqualValues(t, expected.Status, actual.Status, "Status "+baseCase)

			}

		}

	}

}

func TestDockerService_Start(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	var manager = endly.New()
	var useCases = []struct {
		baseDir  string
		Request  *ssh.StartRequest
		Expected *ssh.ContainerInfo
		Error    string
	}{
		{
			"test/start/linux",
			&ssh.StartRequest{
				BaseRequest: &ssh.BaseRequest{
					Target: target,
					Name:   "db1",
				},
			},
			&ssh.ContainerInfo{
				ContainerID: "b5bcc949f075",
				Port:        "0.0.0.0:3306->3306/tcp",
				Command:     "docker-entrypoint...",
				Image:       "mysql:5.6",
				Status:      "up",
				Names:       "db1",
			},
			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		exec.GetReplayService(useCase.baseDir)
		if assert.Nil(t, err) {
			service, err := context.Service(ssh.ServiceID)
			assert.Nil(t, err)

			if assert.Nil(t, err) {
				serviceResponse := service.Run(context, useCase.Request)

				var baseCase = useCase.baseDir + " "
				assert.Equal(t, useCase.Error, serviceResponse.Error, baseCase)

				response, ok := serviceResponse.Response.(*ssh.StartResponse)
				if !ok {
					assert.Fail(t, fmt.Sprintf("process serviceResponse was empty %v %T", baseCase, serviceResponse.Response))
					continue
				}

				var expected = useCase.Expected
				var actual = response

				assert.Equal(t, expected.ContainerID, actual.ContainerID, "ContainerID "+baseCase)
				assert.EqualValues(t, expected.Port, actual.Port, "Port "+baseCase)
				assert.Equal(t, expected.Command, actual.Command, "Command "+baseCase)
				assert.EqualValues(t, expected.Image, actual.Image, "Image "+baseCase)
				assert.EqualValues(t, expected.Names, actual.Names, "Names "+baseCase)
				assert.EqualValues(t, expected.Status, actual.Status, "Status "+baseCase)

			}

		}

	}

}

func TestDockerService_Stop(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	var manager = endly.New()
	var useCases = []struct {
		baseDir  string
		Request  *ssh.StopRequest
		Expected *ssh.ContainerInfo
		Error    string
	}{
		{
			"test/stop/linux",
			&ssh.StopRequest{
				BaseRequest: &ssh.BaseRequest{
					Target: target,
					Name:   "db1",
				},
			},
			&ssh.ContainerInfo{
				ContainerID: "b5bcc949f075",
				Port:        "0.0.0.0:3306->3306/tcp",
				Command:     "docker-entrypoint...",
				Image:       "mysql:5.6",
				Status:      "down",
				Names:       "db1",
			},
			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		exec.GetReplayService(useCase.baseDir)
		if assert.Nil(t, err) {
			service, err := context.Service(ssh.ServiceID)
			assert.Nil(t, err)

			defer context.Close()
			if assert.Nil(t, err) {
				serviceResponse := service.Run(context, useCase.Request)

				var baseCase = useCase.baseDir + " "
				assert.Equal(t, useCase.Error, serviceResponse.Error, baseCase)
				response, ok := serviceResponse.Response.(*ssh.StopResponse)
				if !ok {
					assert.Fail(t, fmt.Sprintf("process serviceResponse was empty %v %T", baseCase, serviceResponse.Response))
					continue
				}

				var expected = useCase.Expected
				var actual = response

				assert.Equal(t, expected.ContainerID, actual.ContainerID, "ContainerID "+baseCase)
				assert.EqualValues(t, expected.Port, actual.Port, "Port "+baseCase)
				assert.Equal(t, expected.Command, actual.Command, "Command "+baseCase)
				assert.EqualValues(t, expected.Image, actual.Image, "Image "+baseCase)
				assert.EqualValues(t, expected.Names, actual.Names, "Names "+baseCase)
				assert.EqualValues(t, expected.Status, actual.Status, "Status "+baseCase)
			}
		}
	}
}

func TestDockerService_Remove(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //

	var manager = endly.New()
	var useCases = []struct {
		baseDir  string
		Request  *ssh.RemoveRequest
		Expected string
		Error    string
	}{
		{
			"test/remove/linux",
			&ssh.RemoveRequest{
				BaseRequest: &ssh.BaseRequest{
					Target: target,
					Name:   "db1",
				},
			},
			"db1",
			"",
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		exec.GetReplayService(useCase.baseDir)
		if assert.Nil(t, err) {
			service, err := context.Service(ssh.ServiceID)
			assert.Nil(t, err)

			defer context.Close()
			if assert.Nil(t, err) {
				serviceResponse := service.Run(context, useCase.Request)

				var baseCase = useCase.baseDir + " "
				assert.Equal(t, useCase.Error, serviceResponse.Error, baseCase)

				response, ok := serviceResponse.Response.(*ssh.RemoveResponse)
				if !ok {
					assert.Fail(t, fmt.Sprintf("process serviceResponse was empty %v %T", baseCase, serviceResponse.Response))
					continue
				}

				var expected = useCase.Expected
				var actual = response
				assert.Equal(t, expected, actual.Stdout, "Command "+baseCase)

			}

		}

	}
}

func TestDockerService_Login(t *testing.T) {

	parent := toolbox.CallerDirectory(3)
	gcrKeyDockerCredentials := path.Join(parent, "test/gcr_key.json")
	keyDockerCredentials := path.Join(parent, "test/key.json")

	credentialFile, err := util.GetDummyCredential()
	assert.Nil(t, err)
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //

	var manager = endly.New()
	var useCases = []struct {
		baseDir          string
		Request          *ssh.LoginRequest
		ExpectedUserName string
		ExpectedStdout   string
		Error            bool
	}{
		{
			"test/login/gcr_key/darwin",
			&ssh.LoginRequest{

				Target:      target,
				Repository:  "us.gcr.io/myproj",
				Credentials: gcrKeyDockerCredentials,
			},
			"_json_key",
			"Login Succeeded",
			false,
		},
		//{
		//	"test/login/gcr/darwin",
		//	&docker.LoginRequest{
		//
		//		Target:     target,
		//		Repository: "us.gcr.io/myproj",
		//		Credentials: keyDockerCredentials,
		//	},
		//	"oauth2accesstoken",
		//	"Login Succeeded",
		//	false,
		//},
		{
			"test/login/std/darwin",
			&ssh.LoginRequest{

				Target:      target,
				Repository:  "repo.com/myproj",
				Credentials: keyDockerCredentials,
			},
			"",
			"",
			true,
		},
	}

	for _, useCase := range useCases {
		var target = useCase.Request.Target
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		exec.GetReplayService(useCase.baseDir)
		if assert.Nil(t, err) {
			service, err := context.Service(ssh.ServiceID)
			defer context.Close()

			if assert.Nil(t, err) {

				serviceResponse := service.Run(context, useCase.Request)

				if useCase.Error {
					assert.True(t, serviceResponse.Error != "")
					serviceResponse = service.Run(context, &ssh.LogoutRequest{
						Target:     useCase.Request.Target,
						Repository: useCase.Request.Repository,
					})
					continue
				}
				if assert.EqualValues(t, "", serviceResponse.Error, useCase.baseDir) {
					response, ok := serviceResponse.Response.(*ssh.LoginResponse)
					if assert.True(t, ok) {
						assert.EqualValues(t, useCase.ExpectedUserName, response.Username)
						assert.EqualValues(t, useCase.ExpectedStdout, response.Stdout)
					}
					serviceResponse = service.Run(context, &ssh.LogoutRequest{
						Target:     useCase.Request.Target,
						Repository: useCase.Request.Repository,
					})
					assert.EqualValues(t, "", serviceResponse.Error)
				}
			}
		}
	}

}

func TestDockerService_Build(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	manager := endly.New()
	context, err := exec.NewSSHReplayContext(manager, target, "test/build/darwin")
	if !assert.Nil(t, err) {
		return
	}
	defer context.Close()
	service, _ := context.Service(ssh.ServiceID)

	response := service.Run(context, &ssh.BuildRequest{
		Target: target,
		Tag: &ssh.Tag{
			Username: "viant",
			Image:    "site_profile_backup",
			Version:  "0.1",
		},
		Path: "/Users/awitas/go/src/github.vianttech.com/etl/site_profile_backup",
	})
	assert.EqualValues(t, "", response.Error)

}

func TestDockerService_Push(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	if err != nil {
		return
	}
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	manager := endly.New()
	var useCases = []struct {
		baseDir string
		Error   bool
	}{
		{
			baseDir: "test/push/error/darwin",
			Error:   true,
		},
		{
			baseDir: "test/push/success/darwin",
			Error:   false,
		},
	}

	for _, useCase := range useCases {
		context, err := exec.NewSSHReplayContext(manager, target, useCase.baseDir)
		if !assert.Nil(t, err) {
			return
		}
		service, _ := context.Service(ssh.ServiceID)
		response := service.Run(context, &ssh.PushRequest{
			Target: target,
			Tag: &ssh.Tag{
				Username: "viant",
				Image:    "site_profile_backup",
				Version:  "0.1",
			},
		})
		if useCase.Error {
			assert.True(t, response.Error != "")
		} else {
			assert.EqualValues(t, "", response.Error)

		}
	}
}

func TestDockerService_Inspect(t *testing.T) {
	credentialFile, err := util.GetDummyCredential()
	if err != nil {
		return
	}
	var target = url.NewResource("scp://127.0.0.1:22/", credentialFile) //
	manager := endly.New()
	context, err := exec.NewSSHReplayContext(manager, target, "test/inspect/darwin")
	if !assert.Nil(t, err) {
		return
	}
	defer context.Close()
	service, _ := context.Service(ssh.ServiceID)
	serviceResponse := service.Run(context, &ssh.InspectRequest{
		BaseRequest: &ssh.BaseRequest{
			Target: target,
			Name:   "site_backup",
		},
	})
	assert.EqualValues(t, "", serviceResponse.Error)
	response, ok := serviceResponse.Response.(*ssh.InspectResponse)
	if assert.True(t, ok) {
		if assert.True(t, response.Stdout != "") {
			assert.NotNil(t, response.Info)
			var aMap = data.NewMap()
			aMap.Put("Output", toolbox.AsSlice(response.Info))
			ip, has := aMap.GetValue("Output[0].NetworkSettings.IPAddress")
			if assert.True(t, has) {
				assert.EqualValues(t, "172.17.0.2", ip)
			}
		}
	}
}

func TestDockerLoginRequest_Validate(t *testing.T) {
	{
		request := &ssh.LoginRequest{}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.LoginRequest{
			Target: url.NewResource("abc"),
		}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.LoginRequest{
			Repository: "abc",
		}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.LoginRequest{
			Repository: "abc",
			Target:     url.NewResource("abc"),
		}
		assert.Nil(t, request.Validate())
	}
}

func Test_DockerBuildRequest_Validate(t *testing.T) {
	{
		request := ssh.BuildRequest{}
		assert.NotNil(t, request.Validate())
	}
	{
		request := ssh.BuildRequest{Target: url.NewResource("abc"), Tag: &ssh.Tag{}}
		assert.NotNil(t, request.Validate())
	}
	{
		request := ssh.BuildRequest{Target: url.NewResource("abc"),
			Arguments: map[string]string{
				"-t": "image:1.0",
			},
			Path: "/",
			Tag:  &ssh.Tag{Image: "abc"}}
		assert.Nil(t, request.Validate())
	}

	{
		request := ssh.BuildRequest{Target: url.NewResource("abc"),
			Path: "/",
			Tag:  &ssh.Tag{Image: "abc"}}
		assert.Nil(t, request.Validate())
	}
	{
		request := ssh.BuildRequest{Target: url.NewResource("abc"),

			Tag: &ssh.Tag{Image: "abc"}}
		assert.NotNil(t, request.Validate())
	}
}

func TestDockerTag_Validate(t *testing.T) {

	{
		request := &ssh.TagRequest{}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.TagRequest{
			Target: url.NewResource("abc"),
		}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.TagRequest{
			Target:    url.NewResource("abc"),
			SourceTag: &ssh.Tag{},
		}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.TagRequest{
			Target:    url.NewResource("abc"),
			SourceTag: &ssh.Tag{},
			TargetTag: &ssh.Tag{},
		}
		assert.NotNil(t, request.Validate())
	}

	{
		request := &ssh.TagRequest{
			Target:    url.NewResource("abc"),
			SourceTag: &ssh.Tag{},
			TargetTag: &ssh.Tag{
				Image: "abc",
			},
		}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.TagRequest{
			Target: url.NewResource("abc"),
			SourceTag: &ssh.Tag{
				Image: "abc",
			},
			TargetTag: &ssh.Tag{},
		}
		assert.NotNil(t, request.Validate())
	}
	{
		request := &ssh.TagRequest{
			Target: url.NewResource("abc"),
			SourceTag: &ssh.Tag{
				Image: "abc",
			},
			TargetTag: &ssh.Tag{
				Image: "abc",
			},
		}
		assert.Nil(t, request.Validate())
	}

}

func TestDockerTag_String(t *testing.T) {
	{
		tag := &ssh.Tag{
			Image: "abc",
		}
		assert.EqualValues(t, "abc", tag.String())
	}
	{
		tag := &ssh.Tag{
			Image:   "abc",
			Version: "latest",
		}
		assert.EqualValues(t, "abc:latest", tag.String())
	}
	{
		tag := &ssh.Tag{
			Registry: "reg.org",
			Image:    "abc",
			Version:  "latest",
		}
		assert.EqualValues(t, "reg.org/abc:latest", tag.String())
	}
	{
		tag := &ssh.Tag{
			Username: "reg.org",
			Image:    "abc",
			Version:  "latest",
		}
		assert.EqualValues(t, "reg.org/abc:latest", tag.String())
	}
}
