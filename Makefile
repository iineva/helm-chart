
all: check lint package index commit

check:
	if [[ -n "$(shell git status -s)" ]]; then echo "modified/untracked"; exit 1; fi

lint:
	helm lint helm-chart/*

package:
	cd package && helm package ../helm-chart/* && cd -
.PHONY:package

index:
	helm repo index --url https://iineva.github.io/helm-chart/ --merge index.yaml .

commit:
	git add . && git commit -m "Updated: $(shell date +'%Y-%m-%d %H:%M:%S')"