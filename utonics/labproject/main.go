package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/G-Node/gin-cli/ginclient"
	"github.com/G-Node/gin-cli/git"
	"github.com/G-Node/tonic/tonic"
	"github.com/G-Node/tonic/tonic/form"
	"github.com/G-Node/tonic/tonic/worker"
	"github.com/gogs/go-gogs-client"
)

// labProjectConfig extends the tonic config with fields specific to the
// labproject service.
type labProjectConfig struct {
	*tonic.Config
	TemplateRepo string
}

// lbconfig global configuration for tonic
var lpconfig *labProjectConfig

func main() {
	elems := []form.Element{
		{
			ID:       "laborg",
			Label:    "Lab organisation",
			Name:     "organisation",
			Type:     form.Select,
			Required: true,
		},
		{
			ID:          "projname",
			Label:       "Project name",
			Name:        "project",
			Description: "Must not already exist",
			Required:    true,
			Type:        form.TextInput,
		},
		{
			ID:          "teamname",
			Label:       "Team name",
			Name:        "team",
			Description: "Name of the team the project will belong to. If it does not exist it will be created. If left blank, a new team will be created with the same name as the project.",
			Required:    false,
			Type:        form.TextInput,
		},
		{
			ID:          "title",
			Label:       "Title",
			Name:        "title",
			Description: "Project title",
			Type:        form.TextArea,
			Required:    false,
		},
	}
	page1 := form.Page{
		Description: "Creating a new project will create a new set of repositories based on the lab template and a team for granting access to all project members.",
		Elements:    elems,
	}
	lpform := form.Form{
		Pages:       []form.Page{page1},
		Name:        "Project creation",
		Description: "",
	}
	lpconfig = readConfig("labproject.json")
	tsrv, err := tonic.NewService(lpform, setForm, newProject, *lpconfig.Config)
	if err != nil {
		log.Fatal(err)
	}
	err = tsrv.Start()
	if err != nil {
		log.Fatal(err)
	}
	tsrv.WaitForInterrupt()
	tsrv.Stop()

}

func setForm(f form.Form, botClient, userClient *worker.Client) (*form.Form, error) {
	orgs, err := getAvailableOrgsAndTeams(botClient, userClient)
	if err != nil {
		return &f, err
	}

	orgelem := &f.Pages[0].Elements[0]
	teamelem := &f.Pages[0].Elements[2]
	// Add available org names to ValueList for field
	orgList := make([]string, 0, len(orgs))
	teamList := make([]string, 0)
	for availOrg, availTeams := range orgs {
		orgList = append(orgList, availOrg)
		teamList = append(teamList, availTeams...)
	}
	orgelem.ValueList = orgList
	teamelem.ValueList = teamList

	return &f, nil
}

func newProject(values map[string][]string, botClient, userClient *worker.Client) ([]string, error) {
	orgName := values["organisation"][0] // required
	project := values["project"][0]      // required
	title := ""
	teamName := ""
	if len(values["title"]) > 0 {
		title = values["title"][0]
	}
	if len(values["team"]) > 0 {
		teamName = values["team"][0]
	}
	if teamName == "" {
		// Team name not specified; use project name
		teamName = project
	}

	msgs := make([]string, 0, 10)

	// verify that the user is a member of the organisation
	orgOK := false
	validOrgs, err := getAvailableOrgsAndTeams(botClient, userClient)
	if err != nil {
		msgs = append(msgs, "Failed to get list of valid orgs")
		return msgs, err
	}
	for validOrg := range validOrgs {
		if validOrg == orgName {
			orgOK = true
			break
		}
	}

	if !orgOK {
		msgs = append(msgs, fmt.Sprintf("Lab organisation %q is not a valid option. Either user is not a member, or the service is not enabled for that organisation.", orgName))
		return msgs, fmt.Errorf("Invalid organisation %q: Cannot create new project", orgName)
	}

	// TODO: Fail if the team exists and the user is not a member

	// Initialise GIN Client to clone and push repository
	if err := botClient.InitGINClient(); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to initialise GIN Client: %v", err.Error()))
		return msgs, err
	}

	// Create temporary directory for cloning
	tempDirName, err := ioutil.TempDir("", "tonic-clone")
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to create temporary clone directory: %v", err.Error()))
		return msgs, err
	}

	// Clone repository
	msgs = append(msgs, fmt.Sprintf("Cloning template repository %s", lpconfig.TemplateRepo))
	if err := botClient.CloneRepo(lpconfig.TemplateRepo, tempDirName); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to clone repository %q: %v", lpconfig.TemplateRepo, err.Error()))
		return msgs, err
	}

	// CD into new clone
	repoName := strings.Split(lpconfig.TemplateRepo, "/")[1]
	localRepoPath := filepath.Join(tempDirName, repoName)
	origdir, err := os.Getwd()
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to get current working directory: %s", err.Error()))
		return msgs, err
	}
	defer os.Chdir(origdir)
	if err = os.Chdir(localRepoPath); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to change to newly cloned path %q: %s", localRepoPath, err.Error()))
		return msgs, err
	}

	remoteName := "newproject"
	createAndSetRemote := func(name string) error {
		repoOpt := gogs.CreateRepoOption{
			Name:        name,
			Description: title,
			Private:     true,
			AutoInit:    false,
			Readme:      "Default",
		}
		// Create project repository
		msgs = append(msgs, fmt.Sprintf("Creating %s/%s", orgName, repoOpt.Name))
		repo, err := botClient.CreateOrgRepo(orgName, repoOpt)
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to create repository: %v", err.Error()))
			return err
		}
		msgs = append(msgs, fmt.Sprintf("Repository created: %s", repo.FullName))

		// Add new remote
		remoteURL := fmt.Sprintf("%s/%s/%s", botClient.GIN.GitAddress(), orgName, repoOpt.Name)
		msgs = append(msgs, fmt.Sprintf("Preparing to push template to new project (adding remote): %s", remoteURL))
		if err := git.RemoteAdd(remoteName, remoteURL); err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to add remote: %s", err.Error()))
			return err
		}
		msgs = append(msgs, fmt.Sprintf("Added new remote: %s [%s]", remoteName, remoteURL))
		// Set it as default
		if err := ginclient.SetDefaultRemote(remoteName); err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to set default remote %q: %s", remoteName, err.Error()))
			return err
		}
		return nil
	}

	mainRepo := fmt.Sprintf("%s.main", project)
	if err := createAndSetRemote(mainRepo); err != nil {
		return msgs, err
	}

	// Clone submodules
	msgs = append(msgs, "Cloning submodules")
	initCmd := git.Command("submodule", "init")
	if stdout, stderr, err := initCmd.OutputError(); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to init submodules: %s - %s", string(stdout), string(stderr)))
		return msgs, err
	}
	updCmd := git.Command("submodule", "update")
	if stdout, stderr, err := updCmd.OutputError(); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to update submodules: %s - %s", string(stdout), string(stderr)))
		return msgs, err
	}

	submoduleForEach := func(args ...string) {
		cmdstr := strings.Join(args, " ")
		cmd := git.Command("submodule", "foreach", cmdstr)
		stdout, stderr, err := cmd.OutputError()
		if err != nil {
			errmsg := fmt.Sprintf("Failed to run command %q in all submodules: %s", cmdstr, err.Error())
			msgs = append(msgs, errmsg)
		}
		fmt.Printf(string(stdout))
		fmt.Printf(string(stderr))
	}
	submoduleForEach("git", "checkout", "master") // TODO: find default branch instead
	submoduleForEach("git", "pull")

	submodules, err := parseGitModules(".")
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to parse .gitmodules: %v", err.Error()))
		return msgs, err
	}

	newSubmodules := make(map[string]*module, len(submodules))
	for smName, submodule := range submodules {
		os.Chdir(submodule.path)
		smName = strings.ReplaceAll(smName, "/", "_") // don't allow / in submodule names
		if err := createAndSetRemote(project + "." + smName); err != nil {
			return msgs, err
		}
		// use relative URLs
		url := fmt.Sprintf("../%s.%s", project, smName)
		newSubmodules[smName] = &module{
			path:   submodule.path,
			url:    url,
			branch: submodule.branch,
		}
		os.Chdir(localRepoPath)
	}

	//	// check if common submodule exists
		//var common *module
		//commonsName := "labcommons"
	//	repoinfo, err := botClient.GetRepo(orgName, commonsName)
	//	if err == nil {
	//		common = &module{
	//			path:   commonsName,
	//			url:    fmt.Sprintf("../%s", commonsName),
	//			branch: repoinfo.DefaultBranch,
	//		}
	//		msgs = append(msgs, fmt.Sprintf("Adding common submodule %q", commonsName))
	//	} else {
	//		msgs = append(msgs, fmt.Sprintf("Common repository %s/%s not found: %s", orgName, commonsName, err.Error()))
	//	}

	// Write back updated .gitmodules file
	msgs = append(msgs, "Updating .gitmodules configuration")
	if err := writeGitModules(localRepoPath, newSubmodules, common); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to write .gitmodules file: %s", err.Error()))
		return msgs, err
	}

	

	// 	msgs = append(msgs, "submodule content cannot be initialised and therefore pushed, yet. please initialise with synchronisation script.")

	parentURL := fmt.Sprintf("%s/%s/%s", botClient.GIN.WebAddress(), orgName, mainRepo)
	for _, submodule := range submodules {
		os.Chdir(submodule.path)
		msgs = append(msgs, "Initialising submodule")
		if err := initSubmodule(botClient); err != nil {
			msgs = append(msgs, fmt.Sprintf("Init failed: %s", err.Error()))
			return msgs, err
		}

		msgs = append(msgs, "Writing link to parent in submodule README(s)")
		if err := linkToParent(parentURL); err != nil {
			msgs = append(msgs, fmt.Sprintf("Init failed: %s", err.Error()))
			return msgs, err
		}

		// Commit changes (update README(s) in submodule)
		if err := commit(botClient, []string{"."}, "Add parent repo URLs to README files"); err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to commit README changes: %s", err.Error()))
			return msgs, err
		}

		msgs = append(msgs, "Uploading submodule to new project repository")
		if err := uploadProjectRepository(botClient, remoteName); err != nil {
			msgs = append(msgs, fmt.Sprintf("Upload failed: %s", err.Error()))
			return msgs, err
		}
		os.Chdir(localRepoPath)
	}
	
	// Commit changes (update .gitmodules, gin commit equivalent)
	if err := commit(botClient, []string{".gitmodules"}, "Configure submodules"); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to commit .gitmodules changes: %s", err.Error()))
		return msgs, err
	}
	
	//	updCmd := git.Command("submodule", "update")
	//	if stdout, stderr, err := updCmd.OutputError(); err != nil {
	//		msgs = append(msgs, fmt.Sprintf("Failed to update submodules (2): %s - %s", string(stdout), string(stderr)))
	//		return msgs, err
	//	}
	
	// Commit changes (update submodule to latest commit, git commit equivalent)
	updCmdc := git.Command("commit . ", "-m 'commit latest subm version'")
	if stdout, stderr, err := updCmdc.OutputError(); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to update parent repos: %s - %s", string(stdout), string(stderr)))
		return msgs, err
	}

	// Push
	msgs = append(msgs, "Uploading template to new project repository")
	if err := uploadProjectRepository(botClient, remoteName); err != nil {
		msgs = append(msgs, fmt.Sprintf("Upload failed: %s", err.Error()))
		return msgs, err
	}
	

	orgTeams, err := botClient.ListTeams(orgName)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to list teams for org: %s", orgName))
		return msgs, err
	}

	// Check if Team exists
	var team *gogs.Team
	for _, orgTeam := range orgTeams {
		if orgTeam.Name == teamName {
			team = orgTeam
			msgs = append(msgs, fmt.Sprintf("Team %s exists. Skipping team creation.", teamName))
			break
		}
	}

	if team == nil {
		// Create Team
		// TODO: Use non admin command when it becomes available
		msgs = append(msgs, fmt.Sprintf("Creating team %s/%s", orgName, project))
		team, err = botClient.AdminCreateTeam(orgName, gogs.CreateTeamOption{Name: teamName, Description: title, Permission: "admin"})
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to create team: %s", err.Error()))
			return msgs, err
		}
		msgs = append(msgs, fmt.Sprintf("Team created: %s", team.Name))

		user, err := userClient.GetSelfInfo()
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to retrieve user info: %s", err.Error()))
			return msgs, err
		}

		// Add User to Team
		msgs = append(msgs, fmt.Sprintf("Adding user %q to team %q", user.Login, team.Name))
		botClient.AdminAddTeamMembership(team.ID, user.Login)
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to add user: %s", err.Error()))
			return msgs, err
		}
	}

	// Add Repositories to Team
	msgs = append(msgs, fmt.Sprintf("Adding repositories %q to team %q", mainRepo+" and others", team.Name))
	if err := botClient.AdminAddTeamRepository(team.ID, mainRepo); err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to add repository %q to team: %s", project, err.Error()))
		return msgs, err
	}
	for smName := range submodules {
		repoName := project + "." + strings.ReplaceAll(smName, "/", "_")
		if err := botClient.AdminAddTeamRepository(team.ID, repoName); err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to add repository %q to team: %s", repoName, err.Error()))
			return msgs, err
		}
	}

	return msgs, nil
}

// getAvailableOrgsAndTeams returns a map of organisation names that the user
// and bot both belong to, each mapped to a list of organisation teams that the
// user belongs to.
func getAvailableOrgsAndTeams(botClient, userClient *worker.Client) (map[string][]string, error) {
	// An org is available for management on the service if the user is a
	// member and the bot is an owner or admin.
	botOrgs, err := botClient.ListMyOrgs()
	if err != nil {
		return nil, err
	}

	// get orgs where the bot has admin access
	adminOrgs := make(map[int64]gogs.Organization, len(botOrgs))
	for _, botOrg := range botOrgs {
		teams, err := botClient.ListTeams(botOrg.UserName)
		if err != nil {
			return nil, err
		}
		for _, team := range teams {
			if team.Permission == "admin" || team.Permission == "owner" {
				adminOrgs[botOrg.ID] = *botOrg
			}
		}
	}

	userOrgs, err := userClient.ListMyOrgs()
	if err != nil {
		return nil, err
	}

	validOrgTeams := make(map[string][]string)
	validOrgs := make([]gogs.Organization, 0, len(userOrgs))
	for _, userOrg := range userOrgs {
		if _, ok := adminOrgs[userOrg.ID]; ok {
			validOrgs = append(validOrgs, *userOrg)
			validOrgTeams[userOrg.UserName] = nil
			orgTeams, err := userClient.ListTeams(userOrg.UserName)
			if err != nil {
				// couldn't get teams; assume user has none in this org
				continue
			}
			teams := make([]string, 0, len(orgTeams))
			for _, team := range orgTeams {
				teams = append(teams, team.Name)
			}
			validOrgTeams[userOrg.UserName] = teams
		}
	}

	return validOrgTeams, nil
}

func readConfig(filename string) *labProjectConfig {
	confFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	confData, err := ioutil.ReadAll(confFile)
	if err != nil {
		log.Fatal(err)
	}

	config := new(labProjectConfig)
	if err := json.Unmarshal(confData, config); err != nil {
		log.Fatal(err)
	}

	if config.Config == nil {
		config.Config = new(tonic.Config)
	}

	// Set defaults for any unset values
	if config.CookieName == "" {
		config.CookieName = "utonic-labproject"
		log.Printf("[config] Setting default cookie name: %s", config.CookieName)
	}
	if config.Port == 0 {
		config.Port = 3000
		log.Printf("[config] Setting default port: %d", config.Port)
	}
	if config.DBPath == "" {
		config.DBPath = "./labproject.db"
		log.Printf("[config] Setting default dbpath: %s", config.DBPath)
	}

	// Warn about unset values with no defaults
	unset := make([]string, 0, 5)
	if config.GIN.Web == "" {
		unset = append(unset, "gin.web")
	}
	if config.GIN.Git == "" {
		unset = append(unset, "gin.git")
	}
	if config.GIN.Username == "" {
		unset = append(unset, "gin.username")
	}
	if config.GIN.Password == "" {
		unset = append(unset, "gin.password")
	}
	if config.TemplateRepo == "" {
		unset = append(unset, "templaterepo")
	}
	if len(unset) > 0 {
		log.Printf("WARNING: The following configuration options are unset and have no defaults: %s", strings.Join(unset, ", "))
	}
	return config
}

func commit(botClient *worker.Client, paths []string, msg string) error {
	// Set local git config
	if err := git.SetGitUser(botClient.GIN.Username, botClient.GIN.Username+"@tonic"); err != nil {
		return err
	}
	addchan := make(chan git.RepoFileStatus)
	go git.Add(paths, addchan)
	for stat := range addchan {
		log.Print(stat)
		if stat.Err != nil {
			return stat.Err
		}
	}
	return git.Commit(msg)
}

func initSubmodule(botClient *worker.Client) error {
	// Set local git config
	if err := git.SetGitUser(botClient.GIN.Username, botClient.GIN.Username+"@tonic"); err != nil {
		return err
	}

	return botClient.GIN.InitDir(false)
}

func linkToParent(parentURL string) error {
	// Add a link to the parent repository in the submodule's README (glob for all files starting with README)
	files, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	parentText := fmt.Sprintf("\n\n%s is the parent directory\n\n", parentURL)
	for _, file := range files {
		if file.IsDir() {
			// ignore
			continue
		}

		if strings.HasPrefix(file.Name(), "README") {
			// append parent name
			readme, err := os.OpenFile(file.Name(), os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer readme.Close()

			if _, err := readme.WriteString(parentText); err != nil {
				return err
			}
		}
	}
	return nil
}

func uploadProjectRepository(botClient *worker.Client, remote string) error {
	// Set local git config
	if err := git.SetGitUser(botClient.GIN.Username, botClient.GIN.Username+"@tonic"); err != nil {
		return err
	}

	uploadchan := make(chan git.RepoFileStatus)
	go botClient.GIN.Upload([]string{}, []string{remote}, uploadchan)
	for stat := range uploadchan {
		log.Print(stat)
		if stat.Err != nil {
			return stat.Err
		}
	}
	return nil
}

type module struct {
	path   string
	url    string
	branch string
}

// parseGitModules reads .gitmodules and returns a map of the configured
// submodules with their URLs and paths.
func parseGitModules(repoPath string) (map[string]*module, error) {
	gmFilePath := filepath.Join(repoPath, ".gitmodules")
	gitmodulesFile, err := os.Open(gmFilePath)
	if err != nil {
		return nil, err
	}
	modulesText, err := ioutil.ReadAll(gitmodulesFile)
	if err != nil {
		return nil, err
	}
	modules := make(map[string]*module)
	lines := bytes.Split(modulesText, []byte("\n"))
	curname := ""
	nameRE := regexp.MustCompile(`\[submodule "(.*)"\]`)
	pathRE := regexp.MustCompile(`path = (.*)`)
	urlRE := regexp.MustCompile(`url = (.*)`)
	branchRE := regexp.MustCompile(`branch = (.*)`)
	for _, line := range lines {
		if match := nameRE.FindSubmatch(line); match != nil {
			name := string(match[1])
			modules[name] = new(module)
			curname = name
		} else if match := pathRE.FindSubmatch(line); match != nil {
			path := string(match[1])
			modules[curname].path = path
		} else if match := urlRE.FindSubmatch(line); match != nil {
			url := string(match[1])
			modules[curname].url = url
		} else if match := branchRE.FindSubmatch(line); match != nil {
			branch := string(match[1])
			modules[curname].branch = branch
		}
	}
	return modules, nil
}

// writeGitModules writes back the .gitmodules file.
func writeGitModules(repoPath string, modules map[string]*module, common *module) error {
	gmFilePath := filepath.Join(repoPath, ".gitmodules")
	gitmodulesFile, err := os.Create(gmFilePath)
	if err != nil {
		return err
	}

	for smName, submodule := range modules {
		if submodule != nil {
			if _, err := gitmodulesFile.WriteString(composeModuleBlock(smName, *submodule)); err != nil {
				return err
			}
		}
	}

	if common != nil {
		if _, err := gitmodulesFile.WriteString(composeModuleBlock("common", *common)); err != nil {
			return err
		}
	}
	return nil
}

func composeModuleBlock(name string, mod module) string {
	headerLine := fmt.Sprintf("[submodule %q]\n", name)
	pathLine := fmt.Sprintf("\tpath = %s\n", mod.path)
	urlLine := fmt.Sprintf("\turl = %s\n", mod.url)
	branchLine := ""
	if mod.branch != "" { // optional
		branchLine = fmt.Sprintf("\tbranch = %s\n", mod.branch)
	}

	return headerLine + pathLine + urlLine + branchLine
}
