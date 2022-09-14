# cumin

### what is cumin
```
Cumin (/ˈkjuːmɪn/ or US: /ˈkuːmɪn/, or /ˈkʌmɪn/)
Cumin is a spice made from the dried seed of a plant known as Cuminum cyminum.
---
जीरा
/jīrā/
In Indian English, jeera is the same as cumin.
/jira, jIrA, jeeraa, jīrā/
```
cumin is a spice, but this tool is spicier :)

`cumin` clones a github issue to your jira board, so you can track
everything in one place.

### setup

- create a github [personal access token](https://github.com/settings/tokens)
  and store it in the environment variable `GITHUB_TOKEN`
- create a [jira personal access token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) and store it in the 
  environment variable `JIRA_TOKEN`

### usage

```console
cumin clone https://github.com/tektoncd/operator/issues/763 \
--add-to-current-sprint \
--assignee concaf \
--story-points 5 \
--type Bug \
--base https://issues.redhat.com \
--board-id 5310 \
--project SRVKP

2022/09/14 13:24:22 found github token set in environment variable GITHUB_TOKEN, using...
2022/09/14 13:24:22 github client generation successful
2022/09/14 13:24:22 jira client generation successful
2022/09/14 13:24:22 fetching github issue...
2022/09/14 13:24:23 github issue fetch successful
2022/09/14 13:24:23 jira issue title: Enable image signing (with chains) in Tektoncd Operator release pipeline
2022/09/14 13:24:23 the issue will be added to the current sprint
2022/09/14 13:24:23 found one active sprint: Pipelines Sprint 224
2022/09/14 13:24:23 creating jira issue now...
2022/09/14 13:24:24 created issue id: SRVKP-2486
2022/09/14 13:24:24 created issue: https://issues.redhat.com/browse/SRVKP-2486
```

### command structure

```console
clone an issue from github to jira
usage:

cumin clone https://github.com/concaf/cumin/issues/2455 \
--project SRVKP \
--label "imported-from-github" \
--label "groomable" \
--type story \
--fix-version "Pipelines 1.10" \
--priority critical \
--assignee concaf \
--add-to-current-sprint \
--story-points 5 \
--title "this cool upstream issue"

Usage:
  cumin clone [flags]

Flags:
      --add-to-current-sprint      add the jira issue to current sprint?
  -a, --assignee string            assignee to set on the jira issue
      --fix-versions stringArray   fixVersion(s) to set on the jira issue
  -h, --help                       help for clone
  -l, --labels stringArray         labels to add to the jira issue
      --priority string            priority to set on the jira issue, e.g. major, critical, blocker, etc
      --story-points int           story points to add to the jira issue
      --title string               override the title in the jira issue
  -t, --type string                type of the jira issue, e.g. story, bug, etc (default "story")

Global Flags:
      --base string      jira base url
      --board-id int     jira board id
  -p, --project string   jira project to clone the issue into
```
