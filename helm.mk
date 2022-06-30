HELM                        ?= helm
HELM_CHART_NAME             ?= backup-operator
HELM_CHART_DIR 		        ?= charts/backup-operator
HELM_VALUES_FILE            ?= charts/backup-operator/values.yaml
HELM_OUTPUT_DIR             ?= tmp/helm
HELM_REGISTRY               ?= https://finleap-connect.github.io/charts
HELM_REGISTRY_ALIAS         ?= finleap-connect
HELM_RELEASE                ?= vaop

.PHONY: template-clean dependency-update install uninstall template docs

##@ Helm

helm-clean: ## clean up templated helm charts
	@rm -Rf $(HELM_OUTPUT_DIR)

helm-dep: ## update helm dependencies
	@$(HELM) dep update $(HELM_CHART_DIR)

helm-lint: ## lint helm chart
	@$(HELM) lint $(HELM_CHART_DIR)

helm-install-from-repo: ## install helm chart from build artifact
	@$(HELM) repo update
	@$(HELM) upgrade --install $(HELM_RELEASE) $(HELM_REGISTRY_ALIAS)/$(HELM_CHART_NAME) --namespace $(KUBE_NAMESPACE) --version $(VERSION) --values $(HELM_VALUES_FILE) --skip-crds

helm-uninstall: ## uninstall helm chart
	@$(HELM) uninstall $(HELM_RELEASE) --namespace $(KUBE_NAMESPACE)

helm-template: helm-clean ## template helm chart
	@mkdir -p $(HELM_OUTPUT_DIR)
	@$(HELM) template $(HELM_RELEASE) $(HELM_CHART_DIR) --namespace $(KUBE_NAMESPACE) --values $(HELM_VALUES_FILE) --output-dir $(HELM_OUTPUT_DIR) --include-crds

helm-add-finleap: ## add finleap helm chart repo
	@$(HELM) repo add $(HELM_REGISTRY_ALIAS) "$(HELM_REGISTRY)"

helm-set-version-all:
	@find $(HELM_CHART_DIR) -name 'Chart.yaml' -exec $(YQ) e --inplace '.version = "$(VERSION)"' {} \;
	@find $(HELM_CHART_DIR) -name 'Chart.yaml' -exec $(YQ) e --inplace '.appVersion = "$(VERSION)"' {} \;
	@find $(HELM_CHART_DIR) -name 'Chart.yaml' -exec $(YQ) e --inplace '(.dependencies.[].version | select(. == "0.0.1-local")) |= "$(VERSION)"' {} \;

helm-docs: ## update the auto generated docs of all helm charts
	@docker run --rm --volume "$(PWD):/helm-docs" -u $(shell id -u) jnorwood/helm-docs:v1.4.0 --template-files=./README.md.gotmpl
