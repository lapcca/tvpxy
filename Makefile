# Paths to tools needed in dependencies
GO := $(shell which go)

# Build flags
BUILD_FLAGS = -ldflags "-s -w -extldflags '-static'" 

# Output directory
BUILD_DIR := build

# Output binary name
OUTPUT_NAME := tvpxy

# Targets
all: clean dependencies build

build: mkdir
	@echo "Building main.go with CGO enabled..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO} build ${BUILD_FLAGS} -tags netgo -o ${BUILD_DIR}/${OUTPUT_NAME} main.go
	@cp -f ${BUILD_DIR}/${OUTPUT_NAME} ..

build-openwrt-amd64: mkdir
	@echo "Building main.go for OpenWrt (amd64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO} build ${BUILD_FLAGS} -tags netgo -o ${BUILD_DIR}/openwrt_amd64_${OUTPUT_NAME} main.go

build-openwrt-arm: mkdir
	@echo "Building main.go for OpenWrt (arm)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm ${GO} build ${BUILD_FLAGS} -tags netgo -o ${BUILD_DIR}/openwrt_arm_${OUTPUT_NAME} main.go

build-openwrt-mips: mkdir
	@echo "Building main.go for OpenWrt (mips)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=mips ${GO} build ${BUILD_FLAGS} -tags netgo -o ${BUILD_DIR}/openwrt_mips_${OUTPUT_NAME} main.go

dependencies:
	@echo "Checking dependencies..."
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

mkdir:
	@echo "Creating build directory..."
	mkdir -p ${BUILD_DIR}
	

clean:
	@echo "Cleaning up..."
	@${GO} mod tidy
	@${GO} clean
	@rm -fr $(BUILD_DIR)
	@rm -rf ${OUTPUT_NAME}
