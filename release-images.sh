#!/bin/bash

# Multi-Carrier Shipping Platform - Docker Image Release Tool
# Automates building, tagging, and publishing Docker images for all microservices.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Default Configuration
REGISTRY_DEFAULT="ghcr.io/kavix/multi-carrier-shipping-golang"
SERVICES=(
    "api-gateway"
    "shipment-service"
    "carrier-integration-service"
    "rate-comparison-service"
    "label-generation-service"
    "tracking-service"
    "address-validation-service"
    "billing-service"
    "return-service"
    "notification-service"
)

# Parse Arguments
TAG=""
REGISTRY=""
LOCAL_ONLY=false
CREATE_GH_RELEASE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --local)
            LOCAL_ONLY=true
            shift
            ;;
        --github-release)
            CREATE_GH_RELEASE=true
            shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        *)
            if [ -z "$TAG" ]; then
                TAG="$1"
                shift
            else
                print_error "Unknown argument: $1"
                exit 1
            fi
            ;;
    esac
done

# If tag is empty, check environment variable TAG
if [ -z "$TAG" ]; then
    TAG="${TAG:-}"
fi

# Fallback: prompt for tag if still empty
if [ -z "$TAG" ]; then
    print_info "No tag specified. Attempting to get latest Git tag..."
    LATEST_GIT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || true)
    if [ -n "$LATEST_GIT_TAG" ]; then
        read -p "Use latest Git tag ($LATEST_GIT_TAG)? [Y/n]: " use_latest
        use_latest=${use_latest:-Y}
        if [[ $use_latest =~ ^[Yy]$ ]]; then
            TAG="$LATEST_GIT_TAG"
        fi
    fi
fi

if [ -z "$TAG" ]; then
    read -p "Enter release tag (e.g., v1.0.0): " input_tag
    TAG="$input_tag"
fi

if [ -z "$TAG" ]; then
    print_error "Release tag is required."
    exit 1
fi

if [ -z "$REGISTRY" ]; then
    REGISTRY="$REGISTRY_DEFAULT"
fi

# Sanitize tag name
TAG=$(echo "$TAG" | tr -d '[:space:]')

print_header "Docker Release Tool ($TAG)"
print_info "Target Registry: $REGISTRY"

# Check Docker
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker."
    exit 1
fi
print_success "Docker is running"

# Build all images
for SERVICE in "${SERVICES[@]}"; do
    DOCKERFILE="./$SERVICE/Dockerfile"
    if [ ! -f "$DOCKERFILE" ]; then
        print_error "Dockerfile not found for service: $SERVICE ($DOCKERFILE)"
        exit 1
    fi
    
    print_header "Building: $SERVICE"
    
    docker build \
        -t "$REGISTRY/$SERVICE:$TAG" \
        -t "$REGISTRY/$SERVICE:latest" \
        -f "$DOCKERFILE" .
        
    print_success "Successfully built and tagged $SERVICE"
done

# Publish images if not local-only
if [ "$LOCAL_ONLY" = false ]; then
    print_header "Publishing Images to $REGISTRY"
    
    # Try logging in if needed
    read -p "Do you want to log in to the registry ($REGISTRY)? [y/N]: " do_login
    if [[ $do_login =~ ^[Yy]$ ]]; then
        # Determine registry host
        REGISTRY_HOST=$(echo "$REGISTRY" | cut -d'/' -f1)
        if [ "$REGISTRY_HOST" = "ghcr.io" ]; then
            print_info "NOTE: To log in to ghcr.io, use your GitHub username and a Personal Access Token (PAT) with 'write:packages' scope as the password."
        fi
        print_info "Logging in to $REGISTRY_HOST..."
        docker login "$REGISTRY_HOST"
    fi
    
    for SERVICE in "${SERVICES[@]}"; do
        print_info "Pushing $SERVICE..."
        docker push "$REGISTRY/$SERVICE:$TAG"
        docker push "$REGISTRY/$SERVICE:latest"
        print_success "Pushed $SERVICE"
    done
    print_success "All images published successfully!"
else
    print_info "Local build complete. Skipped publishing."
fi

# Create GitHub Release if requested
if [ "$CREATE_GH_RELEASE" = true ]; then
    print_header "Creating GitHub Release"
    if ! command -v gh >/dev/null 2>&1; then
        print_error "GitHub CLI (gh) is not installed."
        exit 1
    fi
    
    if ! gh auth status >/dev/null 2>&1; then
        print_error "GitHub CLI is not authenticated. Please run 'gh auth login' first."
        exit 1
    fi
    
    print_info "Creating git tag $TAG..."
    git tag -a "$TAG" -m "Release $TAG" || true
    git push origin "$TAG" || true
    
    print_info "Creating GitHub Release $TAG..."
    gh release create "$TAG" \
        --title "Release $TAG" \
        --notes "Docker images published to $REGISTRY" \
        --draft
        
    print_success "GitHub draft release created successfully!"
fi

print_header "Release Process Completed! 🚀"
