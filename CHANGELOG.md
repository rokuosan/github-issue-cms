# Changelog

## [v1.3.0](https://github.com/rokuosan/github-issue-cms/compare/v1.2.0...v1.3.0) - 2026-11-27

- feat: add OGP image rendering package (pkg/ogimage) by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/158
- feat: add ogimage subcommand for OGP image generation from markdown by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/159
- chore(deps): update dependency goreleaser to v2.17.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/161
- refactor(ogimage): simplify default OGP template by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/163
- feat: integrate OGP image generation into generate command (--with-ogimage) by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/160
- ci: run OGP renderer integration test by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/164
- perf(ogimage): reuse browser during generation by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/165
- Filter issue selection labels from OGP tags by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/167
- fix(ogimage): wait for render readiness by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/166
- feat(ogimage): add live template preview server by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/168
- chore(deps): update actions/deploy-pages action to v5.0.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/176
- chore(deps): update dependency golangci-lint to v2.15.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/173
- chore(deps): update dependency goreleaser to v2.20.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/175
- chore(deps): update dependency golangci/golangci-lint to v2.15.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/174
- chore(deps): update goreleaser/goreleaser-action action to v7.3.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/179
- fix: restrict image tokens to known GitHub hosts by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/151
- fix(deps): update module github.com/google/go-github/v86 to v93 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/180
- chore(deps): update dependency goreleaser to v2.20.3 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/182
- fix(deps): update module github.com/stretchr/testify to v1.12.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/172
- fix(deps): update module github.com/pelletier/go-toml/v2 to v2.4.3 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/138
- ci: adopt tagpr for releases by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/185
- chore: group Go runtime and GitHub Actions updates in Renovate by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/181
- chore: remove broad GitHub Actions update grouping by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/189
- chore(deps): update go version to v1.27.4 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/188
- ci: fix tagpr release triggers and container publishing by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/190

## [v1.2.0](https://github.com/rokuosan/github-issue-cms/compare/v1.1.1...v1.2.0) - 2026-07-26

- refactor: collect and replace GitHub-hosted attachment URLs only by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/126
- chore(deps): update dependency go to v1.26.3 - autoclosed by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/122
- chore(deps): update golang docker tag to v1.26.3 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/123
- fix(deps): update module github.com/google/go-github/v85 to v86 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/124
- fix(deps): update module github.com/google/go-github/v85 to v86 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/127
- feat(config): allow configurable image target URL patterns by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/128
- chore(deps): update goreleaser/goreleaser-action action to v7.2.2 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/130
- chore(deps): update golangci/golangci-lint-action action to v9.2.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/132
- chore(deps): update dependency goreleaser to v2.16.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/133
- chore(deps): update actions/checkout action to v6.0.3 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/134
- chore(deps): update dependency go to v1.26.4 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/135
- chore(deps): update alpine docker tag to v3.24 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/137
- chore(deps): update golang docker tag to v1.26.4 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/136
- fix: log image download authentication failures by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/148
- chore(deps): update actions/checkout action to v7 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/139
- chore(deps): update actions/setup-go action to v7 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/147
- chore(deps): update dependency goreleaser to v2.17.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/143
- chore(deps): update golang docker tag to v1.26.5 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/146
- chore(deps): update goreleaser/goreleaser-action action to v7.2.3 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/141
- chore(deps): update golangci/golangci-lint-action action to v9.3.0 - autoclosed by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/142
- chore(deps): update dependency go to v1.26.5 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/145
- fix: prevent init from overwriting config by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/150
- fix: write downloaded images atomically by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/149
- fix: enforce HTTPS for token-bearing image requests by @rokuosanai in https://github.com/rokuosan/github-issue-cms/pull/152
- docs: note that private attachment images require a classic PAT by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/156
- ci: reuse existing draft release on retry by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/157

## [v1.1.1](https://github.com/rokuosan/github-issue-cms/compare/v1.1.0...v1.1.1) - 2026-05-08

- chore(deps): update dependency golangci-lint to v2.12.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/117
- chore(deps): update dependency golangci/golangci-lint to v2.12.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/118
- chore(deps): update dependency golangci-lint to v2.12.2 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/121
- chore(deps): update dependency golangci/golangci-lint to v2.12.2 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/120
- fix(deps): update module github.com/pelletier/go-toml/v2 to v2.3.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/119
- fix: use front matter date for output path templates by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/125

## [v1.1.0](https://github.com/rokuosan/github-issue-cms/compare/v1.0.0...v1.1.0) - 2026-04-30

- fix: avoid rendering empty custom front matter by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/116
- feat: support filtering issues by label by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/115

## [v1.0.0](https://github.com/rokuosan/github-issue-cms/compare/v0.6.1...v1.0.0) - 2026-04-30

- Update golang-ci by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/42
- chore(deps): update dependency go to 1.24 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/39
- build: update go version to 1.24 by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/44
- chore(deps): update golangci/golangci-lint-action action to v6.5.2 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/43
- fix(deps): update module github.com/spf13/cobra to v1.9.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/38
- fix(deps): update module github.com/spf13/viper to v1.20.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/40
- fix(deps): update module github.com/spf13/viper to v1.20.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/47
- refactor: pkg/converter by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/45
- refactor: replace manual mock with gomock by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/55
- chore(deps): update alpine docker tag to v3.22 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/51
- refactor: improve GitHub Actions workflow for container image builds by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/56
- chore: remove Trivy vulnerability scanner from build workflow by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/59
- chore: update golangci-lint from v1.64.8 to v2.3.0 by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/58
- chore(deps): update docker/metadata-action action to v5 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/57
- docs: add English translations for documentation by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/61
- refactor: set English as default language for documentation by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/62
- fix: refine push condition for Docker image build by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/63
- test: add some tests by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/60
- Add Takumi badge by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/80
- Revert "Add Takumi badge" by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/81
- Pin documented install version to v0.6.1 by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/86
- Restructure generation pipeline and change CLI behavior by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/85
- Update Go toolchain and align CI/Docker configuration by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/87
- Update module github.com/stretchr/testify to v1.11.1 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/70
- Update module github.com/spf13/cobra to v1.10.2 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/72
- Update module github.com/pelletier/go-toml/v2 to v2.3.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/88
- Update actions/attest-build-provenance action to v4 - autoclosed by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/89
- Update alpine Docker tag to v3.23 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/82
- Update module github.com/spf13/viper to v1.21.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/84
- Update actions/checkout action to v6 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/79
- Update golangci/golangci-lint-action action to v9 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/78
- Update actions/setup-go action to v6 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/73
- Improve container image build performance by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/90
- go mod tidy by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/91
- Update actions/configure-pages action to v6 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/92
- Update actions/deploy-pages action to v5 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/93
- Update docker/build-push-action action to v7 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/95
- Update actions/upload-pages-artifact action to v5 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/94
- Update docker/metadata-action action to v6 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/97
- Update docker/login-action action to v4 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/96
- Update docker/setup-qemu-action action to v4 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/99
- Update docker/setup-buildx-action action to v4 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/98
- Update module github.com/google/go-github/v67 to v84 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/100
- add CI check for go mod tidy by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/102
- chore: pin GitHub Actions SHAs in CI workflows by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/104
- Update module github.com/google/go-github/v67 to v85 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/103
- docs: update README GitHub Actions example by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/105
- feat: add version subcommand and inject CLI version in container builds by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/106
- automate GitHub Releases for CLI binaries by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/107
- Update goreleaser/goreleaser-action action to v6.4.0 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/109
- Update goreleaser/goreleaser-action action to v7 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/110
- ci: use GitHub native release notes by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/111
- Refactor config schema to use `output` and add config migration command by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/112
- docs: refresh documentation for v1.0.0 release by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/113

## [v0.6.1](https://github.com/rokuosan/github-issue-cms/compare/v0.6.0...v0.6.1) - 2025-01-25

- docs: Update README.md by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/20
- Add a docker image build workflow by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/22
- Configure golang-ci by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/25
- chore(deps): update actions/configure-pages action to v5 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/27
- chore(deps): update module github.com/google/go-github/v56 to v67 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/28
- chore(deps): update alpine docker tag to v3.21 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/26
- chore(deps): update module github.com/google/go-github/v56 to v67 by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/29
- chore(deps): update module github.com/google/go-github/v56 to v67 - abandoned by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/30
- chore: Configure Renovate by @renovate[bot] in https://github.com/rokuosan/github-issue-cms/pull/31
- 出力時のプレースホルダが Issue 内にある date の値を見るようにした by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/35

## [v0.6.0](https://github.com/rokuosan/github-issue-cms/compare/v0.5.0...v0.6.0) - 2024-07-13

- fix: Fix typo by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/17
- Feature/#13 support page bundle by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/18
- Feature/variable config by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/19

## [v0.5.0](https://github.com/rokuosan/github-issue-cms/compare/v0.4.2...v0.5.0) - 2024-07-01

- Update Go version and create type defined config by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/15
- Feature/#14 output directory by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/16

## [v0.4.2](https://github.com/rokuosan/github-issue-cms/compare/v0.4.1...v0.4.2) - 2024-06-24

## [v0.4.1](https://github.com/rokuosan/github-issue-cms/compare/v0.4.0...v0.4.1) - 2024-06-24

- Add documentation by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/9
- Develop by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/10
- ci: update by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/11
- Develop/docs by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/12

## [v0.4.0](https://github.com/rokuosan/github-issue-cms/compare/v0.3.0...v0.4.0) - 2024-03-30

- Develop by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/8

## [v0.3.0](https://github.com/rokuosan/github-issue-cms/compare/v0.2.4...v0.3.0) - 2024-02-24

- feat: Support custom destination of the articles by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/7

## [v0.2.4](https://github.com/rokuosan/github-issue-cms/compare/v0.2.3...v0.2.4) - 2024-02-23

- fix: Fix the panic out of range by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/6

## [v0.2.3](https://github.com/rokuosan/github-issue-cms/compare/v0.2.2...v0.2.3) - 2023-12-26

## [v0.2.2](https://github.com/rokuosan/github-issue-cms/compare/v0.2.1...v0.2.2) - 2023-12-26

- fix: Support downloading HTML Style image by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/2

## [v0.2.1](https://github.com/rokuosan/github-issue-cms/compare/v0.2.0...v0.2.1) - 2023-12-25

## [v0.2.0](https://github.com/rokuosan/github-issue-cms/compare/v0.1.0...v0.2.0) - 2023-12-22

- Develop by @rokuosan in https://github.com/rokuosan/github-issue-cms/pull/1

## [v0.1.0](https://github.com/rokuosan/github-issue-cms/commits/v0.1.0) - 2023-11-27
