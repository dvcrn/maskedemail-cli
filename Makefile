
GO_SRC := $(wildcard ./*.go ./pkg/*.go)
INSTALL_DIR := ${HOME}/.local/bin
TARGET_BIN := maskedemail-cli

# taken from: https://gist.github.com/grihabor/4a750b9d82c9aa55d5276bd5503829be
# Version Start
USE_VERSION := false
# If this isn't a git repo or the repo has no tags, git describe will return non-zero
ifeq ($(shell git describe > /dev/null 2>&1 ; echo $$?), 0)
	USE_VERSION		   := true
	DESCRIBE           := $(shell git describe --match "v*" --always --tags --dirty)
	DESCRIBE_PARTS     := $(subst -, ,$(DESCRIBE))

	VERSION_TAG        := $(word 1,$(DESCRIBE_PARTS))
	COMMITS_SINCE_TAG  := $(word 2,$(DESCRIBE_PARTS))
	DIRTY              := $(word 4,$(DESCRIBE_PARTS))
	# prepend a dash to dirty if its set
	ifeq ($(DIRTY), dirty)
		DIRTY := -$(DIRTY)
	endif

	VERSION            := $(subst v,,$(VERSION_TAG))
	VERSION_PARTS      := $(subst ., ,$(VERSION))

	MAJOR              := $(word 1,$(VERSION_PARTS))
	MINOR              := $(word 2,$(VERSION_PARTS))
	PATCH              := $(word 3,$(VERSION_PARTS))

	NEXT_MAJOR         := $(shell echo $$(($(MAJOR)+1)))
	NEXT_MINOR         := $(shell echo $$(($(MINOR)+1)))
	NEXT_PATCH          = $(shell echo $$(($(PATCH)+$(COMMITS_SINCE_TAG))))

	ifeq ($(strip $(COMMITS_SINCE_TAG)),)
	CURRENT_VERSION_PATCH := $(MAJOR).$(MINOR).$(PATCH)
	CURRENT_VERSION_MINOR := $(CURRENT_VERSION_PATCH)
	CURRENT_VERSION_MAJOR := $(CURRENT_VERSION_PATCH)
	else
	CURRENT_VERSION_PATCH := $(MAJOR).$(MINOR).$(NEXT_PATCH)
	CURRENT_VERSION_MINOR := $(MAJOR).$(NEXT_MINOR).0
	CURRENT_VERSION_MAJOR := $(NEXT_MAJOR).0.0
	endif

	DATE                = $(shell date +'%Y%m%d')
	TIME                = $(shell date +'%H%M%S')
	COMMIT             := $(shell git rev-parse --short HEAD)
	#AUTHOR             := $(firstword $(subst @, ,$(shell git show --format="%aE" $(COMMIT))))
	#BRANCH_NAME        := $(shell git rev-parse --abbrev-ref HEAD)

	# TAG_MESSAGE         = "$(TIME) $(DATE) $(AUTHOR) $(BRANCH_NAME)"
	# COMMIT_MESSAGE     := $(shell git log --format=%B -n 1 $(COMMIT))

	# CURRENT_TAG_PATCH  := "v$(CURRENT_VERSION_PATCH)"
	# CURRENT_TAG_MINOR  := "v$(CURRENT_VERSION_MINOR)"
	# CURRENT_TAG_MAJOR  := "v$(CURRENT_VERSION_MAJOR)"
endif
# Version End

# set a default build version to PATCH if nothing specified
all: BUILD_VERSION = $(CURRENT_VERSION_PATCH)
all: build

install: ${INSTALL_DIR}/${TARGET_BIN}

${INSTALL_DIR}/${TARGET_BIN}: bin/${TARGET_BIN}
	mkdir -p ${INSTALL_DIR}
	cp -f $< $@

build: bin/${TARGET_BIN}

bin/%: ${GO_SRC}
	mkdir bin/ &> /dev/null || true

ifeq ($(USE_VERSION),true)
# version is based on a git tag "vX.Y.Z" existing
# if such a tag does not exist, empty value will be passed
	@echo "Building version $(BUILD_VERSION)-$(COMMIT)$(DIRTY) of application:"
	go build -ldflags "-X 'main.buildVersion=$(BUILD_VERSION)' -X 'main.buildCommit=$(COMMIT)$(DIRTY)'" -o $@
else
	@echo "Building application:"
	go build -o $@
endif

.PHONY: clean
clean:
	rm -rf bin/

# # --- Version commands ---
# these need to be AFTER the all recipe or otherwise BUILD_VERSION will get set even if those are not specified
.PHONY: version
version: version-patch

.PHONY: version-patch
version-patch: BUILD_VERSION = $(CURRENT_VERSION_PATCH)
version-patch: build

.PHONY: version-minor
version-minor: BUILD_VERSION = $(CURRENT_VERSION_MINOR)
version-minor: build

.PHONY: version-major
version-major: BUILD_VERSION = $(CURRENT_VERSION_MAJOR)
version-major: build
