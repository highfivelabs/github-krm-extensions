# github-krm-extensions

Kubernetes kinds for the GitHub configuration that Terraform's
`integrations/github` provider manages. The only official GitHub operator,
actions-runner-controller, runs Actions runners and manages no GitHub
configuration. So these kinds cover the provider's resources themselves:
repositories, organization settings, teams and rulesets included.

Until 2026-11-26 those four were left to a community operator, Interhyp's
git-hubby. It is young (v0.13, API `v1alpha1`) and mostly one maintainer's
work. Its objects also reset or remove whatever their spec leaves out,
which is unsafe when adopting an estate that already exists. So we built
our own kinds (HFL-777).

By the numbers: `integrations/github` is the #20 Terraform provider by
downloads (421,560,601), and version 6.13.0 has 88 resources. This
repository has **75 kinds**, standing in for **83** of them. The other 5
will not be built ("Not covered", below). Each kind is listed below with
its Terraform equivalent.

**Readiness.** Every kind passes create, update, drift repair and delete
against a fake built from the same descriptors (`TestEveryKindOnFake`).
36 kinds have also passed against the real API, in a test organization on
GitHub's free plan (`hack/live/live.yaml`, `hack/live/records.json`).
`zz_generated.report.md` gives each kind's readiness.

## How it works

The kinds are generated, not hand-written. `hack/gen/vendor.yaml` pins
GitHub's OpenAPI spec; [krm-rest-runtime](https://github.com/highfivelabs/krm-rest-runtime)'s
generator finds every resource the API exposes in a create/read/update/delete
shape, keeps those that stand in for a Terraform resource in the gap list
(`hack/gen/terraform-gaps.json`), and writes:

- `types.yaml`: per kind, its API calls, path parameters and Terraform equivalents
- `config/crds/`: one CRD per kind, its spec the API's request body
- `zz_generated.report.md`: the full list

One generic manager (`cmd/manager`) runs every kind on krm-rest-runtime's
reconciler: it creates the resource, reads it back every 10 minutes and
puts back anything changed outside Kubernetes, and deletes it with the
object. Spec fields keep GitHub's own names, so the API reference
applies field for field. Path parameters are spec fields too.

Hand corrections to the automatic matching (a Terraform type matched to
the wrong API path, or not matched) are in `vendor.yaml`'s `overrides`.
Resources GitHub manages only through GraphQL have no OpenAPI description:
each is a hand-written file in `hack/gen/graphql/` (its documents, as the
Terraform provider sends them, and how the spec fills their variables),
naming its sources at the top; `vendor.yaml`'s `graphql` names the
endpoint, `/graphql`. REST resources no OpenAPI item describes as one
(a branch: made on `.../git/refs`, read at `.../git/ref/heads/<branch>`),
or that the spec leaves out (EMU group mappings), and the whole-set kinds,
are hand-written files in `hack/gen/kinds/` in the same format, their
calls REST.

## Secrets

GitHub wants secret values encrypted with the repository's (or
organization's, or environment's) public key. Give the plain value as
`spec.secretRefs.value`; the manager fetches the public key, encrypts
the value (libsodium sealed box) and sends `encrypted_value` and `key_id`.

## Install

```sh
kubectl kustomize config/default | KO_DOCKER_REPO=<registry> ko resolve -f - | kubectl apply --server-side -f -
```

Then, in each namespace that manages GitHub, a Secret with
`token` and a Connection named `default`
(`config/samples/connection.yaml`). Objects use it unless
`spec.connectionRef.name` names another. Secret values (tokens, passwords,
secret values) go in a Kubernetes Secret, not the spec:
`spec.secretRefs: {<field>: {name, key}}` fills that body field.

To run only some API groups, install those groups' CRDs from `config/crds`
and pass `--groups` to the manager.

## Kinds

| Kind | Terraform equivalent | Shape | Read |
| --- | --- | --- | --- |
| `actions.github.hfl.dev/v1alpha1` `ActionsEnvironmentSecret` | `github_actions_environment_secret` | upsert | `GET /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsEnvironmentVariable` | `github_actions_environment_variable` | crud | `GET /repos/{owner}/{repo}/environments/{environment_name}/variables/{name}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsHostedRunner` | `github_actions_hosted_runner` | crud | `GET /orgs/{org}/actions/hosted-runners/{hosted_runner_id}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationOidcSubjectClaimCustomizationTemplate` | `github_actions_organization_oidc_subject_claim_customization_template` | settings | `GET /orgs/{org}/actions/oidc/customization/sub` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationPermissions` | `github_actions_organization_permissions` | settings | `GET /orgs/{org}/actions/permissions` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationSecret` | `github_actions_organization_secret` | upsert | `GET /orgs/{org}/actions/secrets/{secret_name}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationSecretRepository` | `github_actions_organization_secret_repositories`, `github_actions_organization_secret_repository` | settings | `GET /orgs/{org}/actions/secrets/{secret_name}/repositories` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationVariable` | `github_actions_organization_variable` | crud | `GET /orgs/{org}/actions/variables/{name}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationVariableRepository` | `github_actions_organization_variable_repositories`, `github_actions_organization_variable_repository` | settings | `GET /orgs/{org}/actions/variables/{name}/repositories` |
| `actions.github.hfl.dev/v1alpha1` `ActionsOrganizationWorkflowPermissions` | `github_actions_organization_workflow_permissions` | settings | `GET /orgs/{org}/actions/permissions/workflow` |
| `actions.github.hfl.dev/v1alpha1` `ActionsRepositoryAccessLevel` | `github_actions_repository_access_level` | settings | `GET /repos/{owner}/{repo}/actions/permissions/access` |
| `actions.github.hfl.dev/v1alpha1` `ActionsRepositoryOidcSubjectClaimCustomizationTemplate` | `github_actions_repository_oidc_subject_claim_customization_template` | settings | `GET /repos/{owner}/{repo}/actions/oidc/customization/sub` |
| `actions.github.hfl.dev/v1alpha1` `ActionsRepositoryPermissions` | `github_actions_repository_permissions` | settings | `GET /repos/{owner}/{repo}/actions/permissions` |
| `actions.github.hfl.dev/v1alpha1` `ActionsRunnerGroup` | `github_actions_runner_group` | crud | `GET /orgs/{org}/actions/runner-groups/{runner_group_id}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsSecret` | `github_actions_secret` | upsert | `GET /repos/{owner}/{repo}/actions/secrets/{secret_name}` |
| `actions.github.hfl.dev/v1alpha1` `ActionsVariable` | `github_actions_variable` | crud | `GET /repos/{owner}/{repo}/actions/variables/{name}` |
| `app.github.hfl.dev/v1alpha1` `AppInstallationRepository` | `github_app_installation_repository` | upsert | `GET /user/installations/{installation_id}/repositories`, the element whose `id` is `{repository_id}` |
| `app.github.hfl.dev/v1alpha1` `AppInstallationRepositorySet` | `github_app_installation_repositories` | set | `GET /user/installations/{installation_id}/repositories`, every element |
| `branch.github.hfl.dev/v1alpha1` `Branch` | `github_branch` | upsert | `GET /repos/{owner}/{repo}/git/ref/heads/{+branch}` |
| `branch.github.hfl.dev/v1alpha1` `BranchDefault` | `github_branch_default` | settings | `GET /repos/{owner}/{repo}` |
| `branch.github.hfl.dev/v1alpha1` `BranchProtection` | `github_branch_protection` | crud | GraphQL `BranchProtectionRule` |
| `branch.github.hfl.dev/v1alpha1` `BranchProtectionV3` | `github_branch_protection_v3` | upsert | `GET /repos/{owner}/{repo}/branches/{branch}/protection` |
| `codespaces.github.hfl.dev/v1alpha1` `CodespacesOrganizationSecret` | `github_codespaces_organization_secret` | upsert | `GET /orgs/{org}/codespaces/secrets/{secret_name}` |
| `codespaces.github.hfl.dev/v1alpha1` `CodespacesOrganizationSecretRepositories` | `github_codespaces_organization_secret_repositories` | settings | `GET /orgs/{org}/codespaces/secrets/{secret_name}/repositories` |
| `codespaces.github.hfl.dev/v1alpha1` `CodespacesSecret` | `github_codespaces_secret` | upsert | `GET /repos/{owner}/{repo}/codespaces/secrets/{secret_name}` |
| `codespaces.github.hfl.dev/v1alpha1` `CodespacesUserSecret` | `github_codespaces_user_secret` | upsert | `GET /user/codespaces/secrets/{secret_name}` |
| `dependabot.github.hfl.dev/v1alpha1` `DependabotOrganizationSecret` | `github_dependabot_organization_secret` | upsert | `GET /orgs/{org}/dependabot/secrets/{secret_name}` |
| `dependabot.github.hfl.dev/v1alpha1` `DependabotOrganizationSecretRepository` | `github_dependabot_organization_secret_repositories`, `github_dependabot_organization_secret_repository` | settings | `GET /orgs/{org}/dependabot/secrets/{secret_name}/repositories` |
| `dependabot.github.hfl.dev/v1alpha1` `DependabotSecret` | `github_dependabot_secret` | upsert | `GET /repos/{owner}/{repo}/dependabot/secrets/{secret_name}` |
| `enterprise.github.hfl.dev/v1alpha1` `EnterpriseActionsPermissions` | `github_enterprise_actions_permissions` | settings | `GET /enterprises/{enterprise}/actions/permissions` |
| `enterprise.github.hfl.dev/v1alpha1` `EnterpriseActionsRunnerGroup` | `github_enterprise_actions_runner_group` | crud | `GET /enterprises/{enterprise}/actions/runner-groups/{runner_group_id}` |
| `enterprise.github.hfl.dev/v1alpha1` `EnterpriseActionsWorkflowPermissions` | `github_enterprise_actions_workflow_permissions` | settings | `GET /enterprises/{enterprise}/actions/permissions/workflow` |
| `enterprise.github.hfl.dev/v1alpha1` `EnterpriseIpAllowListEntry` | `github_enterprise_ip_allow_list_entry` | crud | GraphQL `IpAllowListEntry` |
| `enterprise.github.hfl.dev/v1alpha1` `EnterpriseOrganization` | `github_enterprise_organization` | crud | GraphQL `CreateEnterpriseOrganization`, then `GET /orgs/{id}` |
| `enterprise.github.hfl.dev/v1alpha1` `EnterpriseSecurityAnalysisSettings` | `github_enterprise_security_analysis_settings` | settings | `GET /enterprises/{enterprise}/code_security_and_analysis` |
| `issue.github.hfl.dev/v1alpha1` `Issue` | `github_issue` | crud | `GET /repos/{owner}/{repo}/issues/{issue_number}` |
| `emu.github.hfl.dev/v1alpha1` `EmuGroupMapping` | `github_emu_group_mapping` | upsert | `GET /orgs/{org}/teams/{team_slug}/external-groups`, the element whose `group_id` is `{group_id}` |
| `issue.github.hfl.dev/v1alpha1` `IssueLabel` | `github_issue_label` | crud | `GET /repos/{owner}/{repo}/labels/{name}` |
| `issue.github.hfl.dev/v1alpha1` `IssueLabelSet` | `github_issue_labels` | set | `GET /repos/{owner}/{repo}/labels`, every element |
| `membership.github.hfl.dev/v1alpha1` `Membership` | `github_membership` | upsert | `GET /orgs/{org}/memberships/{username}` |
| `organization.github.hfl.dev/v1alpha1` `OrganizationBlock` | `github_organization_block` | upsert | `GET /orgs/{org}/blocks/{username}` |
| `organization.github.hfl.dev/v1alpha1` `OrganizationRepositoryRole` | `github_organization_custom_role`, `github_organization_repository_role` | crud | `GET /orgs/{org}/custom-repository-roles/{role_id}` |
| `organization.github.hfl.dev/v1alpha1` `OrganizationRole` | `github_organization_role` | crud | `GET /orgs/{org}/organization-roles/{role_id}` |
| `organization.github.hfl.dev/v1alpha1` `OrganizationRoleTeam` | `github_organization_role_team`, `github_organization_role_team_assignment`, `github_organization_security_manager` | upsert | `GET /orgs/{org}/organization-roles/{role_id}/teams`, the element whose `slug` is `{team_slug}` |
| `organization.github.hfl.dev/v1alpha1` `OrganizationRoleUser` | `github_organization_role_user` | upsert | `GET /orgs/{org}/organization-roles/{role_id}/users`, the element whose `login` is `{username}` |
| `organization.github.hfl.dev/v1alpha1` `OrganizationWebhook` | `github_organization_webhook` | crud | `GET /orgs/{org}/hooks/{hook_id}` |
| `release.github.hfl.dev/v1alpha1` `Release` | `github_release` | crud | `GET /repos/{owner}/{repo}/releases/{release_id}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryCollaborator` | `github_repository_collaborator`, `github_repository_collaborators` | upsert | `GET /repos/{owner}/{repo}/collaborators/{username}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryCustomProperty` | `github_repository_custom_property` | upsert | `GET /repos/{owner}/{repo}/properties/values` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryDependabotSecurityUpdate` | `github_repository_dependabot_security_updates` | upsert | `GET /repos/{owner}/{repo}/automated-security-fixes` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryDeployKey` | `github_repository_deploy_key` | crud | `GET /repos/{owner}/{repo}/keys/{key_id}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryDeploymentBranchPolicy` | `github_repository_deployment_branch_policy`, `github_repository_environment_deployment_policy` | crud | `GET /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryEnvironment` | `github_repository_environment` | upsert | `GET /repos/{owner}/{repo}/environments/{environment_name}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryFile` | `github_repository_file` | upsert | `GET /repos/{owner}/{repo}/contents/{path}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryMilestone` | `github_repository_milestone` | crud | `GET /repos/{owner}/{repo}/milestones/{milestone_number}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryPagesSite` | `github_repository_pages` | upsert | `GET /repos/{owner}/{repo}/pages` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryPullRequest` | `github_repository_pull_request` | crud | `GET /repos/{owner}/{repo}/pulls/{pull_number}` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryTopics` | `github_repository_topics` | settings | `GET /repos/{owner}/{repo}/topics` |
| `repository.github.hfl.dev/v1alpha1` `RepositoryVulnerabilityAlerts` | `github_repository_vulnerability_alerts` | upsert | `GET /repos/{owner}/{repo}/vulnerability-alerts` |
| `team.github.hfl.dev/v1alpha1` `TeamMembership` | `github_team_membership` | upsert | `GET /teams/{team_id}/memberships/{username}` |
| `team.github.hfl.dev/v1alpha1` `TeamRepository` | `github_team_repository` | upsert | `GET /teams/{team_id}/repos/{owner}/{repo}` |
| `team.github.hfl.dev/v1alpha1` `TeamSettings` | `github_team_settings` | settings | GraphQL `TeamReviewAssignment` |
| `team.github.hfl.dev/v1alpha1` `TeamSyncGroupMapping` | `github_team_sync_group_mapping` | settings | `GET /orgs/{org}/teams/{team_slug}/team-sync/group-mappings` |
| `user.github.hfl.dev/v1alpha1` `UserGpgKey` | `github_user_gpg_key` | crud | `GET /user/gpg_keys/{gpg_key_id}` |
| `user.github.hfl.dev/v1alpha1` `UserSshKey` | `github_user_ssh_key` | crud | `GET /user/keys/{key_id}` |
| `workflow.github.hfl.dev/v1alpha1` `WorkflowRepositoryPermissions` | `github_workflow_repository_permissions` | settings | `GET /repos/{owner}/{repo}/actions/permissions/workflow` |

GitHub cannot delete an issue or a pull request: deleting an `Issue` or a
`RepositoryPullRequest` object closes it (`PATCH` with `state: closed`), as
Terraform's destroy does, and the closed one stays in the repository.

### GitHub Enterprise Cloud kinds

`EnterpriseActionsPermissions`, `EnterpriseActionsRunnerGroup`,
`EnterpriseActionsWorkflowPermissions`, `EnterpriseOrganization`,
`EnterpriseSecurityAnalysisSettings`, `EmuGroupMapping`,
`OrganizationRepositoryRole`, `OrganizationRole` and `TeamSyncGroupMapping`
use GitHub Enterprise Cloud APIs. They are served by the same host,
`api.github.com`, with the same `token`, so the namespace's usual Connection
works. A classic token needs `admin:enterprise` for the enterprise Actions
settings and security analysis settings, `manage_runners:enterprise` for
`EnterpriseActionsRunnerGroup`, and `admin:org` for the others (GitHub's API
reference for each operation says so); the account must be on Enterprise
Cloud. GitHub's api.github.com spec
leaves these operations out, so `hack/gen/supplement/ghec.yaml` copies them
from GitHub's own GHEC description at the same pinned commit.

`OrganizationRoleTeam`, `OrganizationRoleUser` and `AppInstallationRepository`
are links GitHub makes with a PUT and removes with a DELETE but never reads
singly: the manager reads them by finding the team (by `slug`), user (by
`login`) or repository (by `id`) in the list Terraform reads too, asking for
100 a page. It reads one page, so a role assigned to more than 100 teams or
users, or an installation with more than 100 repositories, can read as
missing. Terraform's `github_organization_security_manager` assigns the
organization's predefined `security_manager` role to a team: an
`OrganizationRoleTeam` with that role's `roleID` (`GET /orgs/{org}/organization-roles`
lists the roles) does the same. `AppInstallationRepository` takes the
repository's numeric ID; its calls (`/user/installations/...`) act as the
token's user, who must have admin access to the repository, so its
Connection needs a user's token, not an app installation token. The role
kinds need an organization owner's token (`admin:org` for a classic one).

`EmuGroupMapping` connects a team of an Enterprise Managed Users
organization to its identity provider group: `PATCH` connects one (in place
of any other), `GET` lists the one there is, `DELETE` disconnects it. These
calls are in GitHub's GHEC description, not the api.github.com one;
`hack/gen/kinds/emu_group_mapping.yaml` describes them.

`EnterpriseOrganization` is made over GraphQL (`createEnterpriseOrganization`,
naming the enterprise by its node ID and the organization's administrators)
and read, changed and deleted over REST at `/orgs/{login}`, as Terraform
does: the reconciler reads a new organization back at once and sets what
the create cannot (its `description` and display `name`). Its
administrators are named on create only; a change to them is not applied.

`Branch` is a Git ref, `refs/heads/<branch>`: made by a `POST` to the
repository's refs at the commit `sha` (Terraform's `source_sha`; Terraform
can look up a source branch's tip instead, which this kind does not), read
at `.../git/ref/heads/<branch>`, deleted at `.../git/refs/heads/<branch>`.
A branch named `feature/x` keeps its slash in those paths. Its head moves
with every push, so `sha` is not compared with it; a changed `sha` moves
the branch there with a fast-forward update (Terraform re-creates it).

`IssueLabelSet` and `AppInstallationRepositorySet` hold a whole set: every
label of a repository (`labels`: name, color, description), every
repository an app installation reaches (`repository_ids`). The manager adds
what is missing, changes what differs and removes what the spec does not
list, one call each, as Terraform's `github_issue_labels` and
`github_app_installation_repositories` do; deleting the object removes
them. GitHub keeps an installation's last repository (its removal answers
422, taken as done). Terraform names repositories; this kind takes their
IDs, which GitHub's calls take.

`BranchDefault` is the `default_branch` field of `/repos/{owner}/{repo}`
alone (`hack/gen/supplement/default-branch.yaml`); like Terraform's
`github_branch_default`, deleting the object leaves the repository's default
branch as it is.

### GraphQL kinds

`BranchProtection`, `EnterpriseIpAllowListEntry` and `TeamSettings` have no
REST API: they post to `https://api.github.com/graphql` (on GitHub
Enterprise Server, `https://<host>/api/graphql`, beside the Connection's
`/api/v3`), with the same `token`. They name what they act on by node ID,
as GitHub's GraphQL does: `BranchProtection`'s `repositoryId` and the actor
lists (who may push, dismiss reviews or bypass), `EnterpriseIpAllowListEntry`'s
`ownerId` (the enterprise), `TeamSettings`' `teamId`; the REST API answers
each object's `node_id`. Terraform takes names and slugs too and looks the
IDs up first. `BranchProtection`'s actor lists are read back as allowance
connections of the actors, so they are sent when the spec changes and not
compared. `TeamSettings` is a settings kind: deleting the object leaves the
team's review settings as they are (Terraform turns delegation off).

`RepositoryFile` writes a file with the contents API: its update and delete
send the blob `sha` a read answers just before (and the delete, the spec's
`message` and `branch`). `spec.branch` is required: the read asks for that
branch's file. `content` (base64) is compared only when the spec changes,
since a read answers it re-encoded. `spec.path` keeps its slashes
(`docs/README.md` in `.../contents/docs/README.md`), as go-github sends it.

### Not covered

<!-- BEGIN generated: not covered (krmgen.py, from vendor.yaml hardBlocks and draftNotes) -->
5 of the 88 Terraform gap types have no kind: 5 are not built (`hardBlocks` in `hack/gen/vendor.yaml`).

| Terraform | Status | Why |
| --- | --- | --- |
| `github_organization_project` | not built | Projects (classic): GitHub has removed its REST API (the pinned REST description has only `projectsV2`; the classic `/projects/...` and `.../projects` paths are gone). |
| `github_project_card` | not built | Projects (classic): GitHub has removed its REST API (the pinned REST description has only `projectsV2`; the classic `/projects/...` and `.../projects` paths are gone). |
| `github_project_column` | not built | Projects (classic): GitHub has removed its REST API (the pinned REST description has only `projectsV2`; the classic `/projects/...` and `.../projects` paths are gone). |
| `github_repository_project` | not built | Projects (classic): GitHub has removed its REST API (the pinned REST description has only `projectsV2`; the classic `/projects/...` and `.../projects` paths are gone). |
| `github_user_invitation_accepter` | not built | An action (accept an invitation: `PATCH /user/repository_invitations/{invitation_id}`), not a resource to keep in a state: the invitation is gone once accepted, and nothing reads or undoes the acceptance. |
<!-- END generated: not covered -->

## For agents (MCP)

`cmd/mcp` serves this repository's kinds over the Model Context Protocol
(stdio), from `types.yaml`, `config/crds` and the live tests' records.
`.mcp.json` registers it, so an agent session opened in this repository
has it without setup. Elsewhere, run it from the repository root:

```sh
go run ./cmd/mcp
```

Its tools: `list_kinds`, `find_kinds_for_terraform`, `describe_kind` (every
field, the REST calls, the live record), `example` (the live-tested spec
where there is one), `validate_manifest` (the CRD schema, as the API server
checks it) and `plan_on_fake` (the controller run against the fake API,
with every REST call it makes). Nothing it does reaches the real API.
krm-rest-runtime's README ("For agents") has the details.

## Regenerate

```sh
hack/gen/generate.sh   # needs python3 + pyyaml, and krm-rest-runtime beside this repo
go test ./...
```
