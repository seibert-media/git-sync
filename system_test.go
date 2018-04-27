package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var pathToServerBinary string
var serverSession *gexec.Session

var _ = BeforeSuite(func() {
	var err error
	pathToServerBinary, err = gexec.Build("github.com/bborbe/git-sync")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

type args map[string]string

func (a args) list() []string {
	var result []string
	for k, v := range a {
		if len(v) == 0 {
			result = append(result, fmt.Sprintf("-%s", k))
		} else {
			result = append(result, fmt.Sprintf("-%s=%s", k, v))
		}
	}
	return result
}

var _ = Describe("git-sync", func() {
	var err error
	It("returns with exitcode != 0 if no parameters have been given", func() {
		serverSession, err = gexec.Start(exec.Command(pathToServerBinary), GinkgoWriter, GinkgoWriter)
		Expect(err).To(BeNil())
		serverSession.Wait(time.Second)
		Expect(serverSession.ExitCode()).NotTo(Equal(0))
	})
	Context("when validating parameters", func() {
		var validargs args
		var targetDirectory string
		BeforeEach(func() {
			targetDirectory = path.Join(os.TempDir(), "git-sync")
			validargs = map[string]string{
				"logtostderr": "",
				"v":           "0",
				"repo":        "http://github.com/bborbe/git-sync.git",
				"one-time":    "",
				"dest":        targetDirectory,
			}
		})
		AfterEach(func() {
			os.RemoveAll(targetDirectory)
		})
		It("returns with exitcode == 0", func() {
			serverSession, err = gexec.Start(exec.Command(pathToServerBinary, validargs.list()...), GinkgoWriter, GinkgoWriter)
			Expect(err).To(BeNil())
			serverSession.Wait(5 * time.Second)
			Expect(serverSession.ExitCode()).To(Equal(0))
			_, err = os.Stat(targetDirectory)
			Expect(os.IsNotExist(err)).To(BeFalse())
		})
		Context("and url parameter", func() {
			var server *ghttp.Server
			BeforeEach(func() {
				server = ghttp.NewServer()
				server.RouteToHandler(http.MethodGet, "/", ghttp.RespondWith(http.StatusOK, "OK"))
			})
			AfterEach(func() {
				serverSession.Interrupt()
				Eventually(serverSession).Should(gexec.Exit())
				server.Close()
			})
			It("calls the url", func() {
				validargs["callback-url"] = server.URL()
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, validargs.list()...), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(5 * time.Second)
				Expect(serverSession.ExitCode()).To(Equal(0))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})
	})
})

func TestSystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "System Test Suite")
}