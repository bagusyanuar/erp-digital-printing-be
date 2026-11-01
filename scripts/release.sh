#!/usr/bin/env bash
set -e

# Colors
CYAN='\033[36m'
YELLOW='\033[33m'
MAGENTA='\033[35m'
GREEN='\033[32m'
RED='\033[31m'
NC='\033[0m' # No Color

echo -e "\n${CYAN}🚀 ========================================${NC}"
echo -e "${CYAN}   Backend Release & Version Bump Tool    ${NC}"
echo -e "${CYAN}========================================${NC}\n"

VERSION_FILE="VERSION"
CURRENT_VERSION=""

# 1. Read current version from VERSION file or latest git tag
if [ -f "$VERSION_FILE" ]; then
  CURRENT_VERSION=$(tr -d ' \n\r' < "$VERSION_FILE")
fi

if [ -z "$CURRENT_VERSION" ]; then
  LATEST_TAG=$(git tag -l 'v*.*.*' 2>/dev/null | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 | sed 's/^v//')
  if [ -n "$LATEST_TAG" ]; then
    CURRENT_VERSION="$LATEST_TAG"
  else
    CURRENT_VERSION="1.0.0"
  fi
fi

echo -e "📌 Current Base Version: ${CYAN}${CURRENT_VERSION}${NC}\n"

# 2. Parse target from arguments or prompt
TARGET="$1"
BUMP_TYPE="$2"

if [ -z "$TARGET" ]; then
  echo "Pilih Target Environment:"
  echo "  1) dev  -> Tag: dev-vX.Y.Z (Memicu Docker CI Dev)"
  echo "  2) prod -> Tag: vX.Y.Z     (Memicu Docker CI/CD Prod)"
  read -p "Pilihan (1/2 atau dev/prod) [default: 1]: " TARGET_ANS
  TARGET_ANS=$(echo "$TARGET_ANS" | tr '[:upper:]' '[:lower:]' | xargs)

  if [ "$TARGET_ANS" = "2" ] || [ "$TARGET_ANS" = "prod" ]; then
    TARGET="prod"
  else
    TARGET="dev"
  fi
fi

TARGET=$(echo "$TARGET" | tr '[:upper:]' '[:lower:]')
if [ "$TARGET" != "dev" ] && [ "$TARGET" != "prod" ]; then
  echo -e "${RED}❌ Target tidak valid! Gunakan 'dev' atau 'prod'.${NC}"
  exit 1
fi

if [ "$TARGET" = "dev" ]; then
  TAG_PREFIX="dev-v"
else
  TAG_PREFIX="v"
fi

echo -e "🎯 Target Environment : ${YELLOW}$(echo "$TARGET" | tr '[:lower:]' '[:upper:]')${NC} (Prefix Git Tag: ${MAGENTA}${TAG_PREFIX}${NC})\n"

# 3. Check candidate tag existence
CANDIDATE_TAG="${TAG_PREFIX}${CURRENT_VERSION}"
IS_CURRENT_TAGGED=false
if git tag -l "$CANDIDATE_TAG" | grep -qx "$CANDIDATE_TAG"; then
  IS_CURRENT_TAGGED=true
fi

# Semver calculation
MAJOR=$(echo "$CURRENT_VERSION" | cut -d. -f1)
MINOR=$(echo "$CURRENT_VERSION" | cut -d. -f2)
PATCH=$(echo "$CURRENT_VERSION" | cut -d. -f3 | sed -E 's/[^0-9].*//')

MAJOR=${MAJOR:-1}
MINOR=${MINOR:-0}
PATCH=${PATCH:-0}

PATCH_PREVIEW="${MAJOR}.${MINOR}.$((PATCH + 1))"
MINOR_PREVIEW="${MAJOR}.$((MINOR + 1)).0"
MAJOR_PREVIEW="$((MAJOR + 1)).0.0"

# 4. Bump type selection
if [ -z "$BUMP_TYPE" ]; then
  echo "Pilih Tipe Version Bump:"
  if [ "$IS_CURRENT_TAGGED" = false ]; then
    echo -e "  0) initial -> Tag: ${MAGENTA}${TAG_PREFIX}${CURRENT_VERSION}${NC} (Versi: ${CURRENT_VERSION})"
  fi
  echo -e "  1) patch   -> Tag: ${MAGENTA}${TAG_PREFIX}${PATCH_PREVIEW}${NC} (Versi: ${PATCH_PREVIEW})"
  echo -e "  2) minor   -> Tag: ${MAGENTA}${TAG_PREFIX}${MINOR_PREVIEW}${NC} (Versi: ${MINOR_PREVIEW})"
  echo -e "  3) major   -> Tag: ${MAGENTA}${TAG_PREFIX}${MAJOR_PREVIEW}${NC} (Versi: ${MAJOR_PREVIEW})"
  echo -e "  4) custom  -> Masukkan versi manual"

  DEFAULT_CHOICE="1"
  PROMPT_RANGE="1-4"
  if [ "$IS_CURRENT_TAGGED" = false ]; then
    DEFAULT_CHOICE="0"
    PROMPT_RANGE="0-4"
  fi

  read -p "Pilihan ($PROMPT_RANGE) [default: $DEFAULT_CHOICE]: " BUMP_ANS
  BUMP_ANS=$(echo "$BUMP_ANS" | tr '[:upper:]' '[:lower:]' | xargs)
  [ -z "$BUMP_ANS" ] && BUMP_ANS="$DEFAULT_CHOICE"

  case "$BUMP_ANS" in
    0|initial)
      BUMP_TYPE="$CURRENT_VERSION"
      ;;
    2|minor)
      BUMP_TYPE="minor"
      ;;
    3|major)
      BUMP_TYPE="major"
      ;;
    4|custom)
      read -p "Masukkan versi baru (contoh: ${CURRENT_VERSION}): " CUSTOM_VER
      CUSTOM_VER=$(echo "$CUSTOM_VER" | xargs)
      if ! echo "$CUSTOM_VER" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+'; then
        echo -e "${RED}❌ Format versi tidak valid (harus semver misal: X.Y.Z)${NC}"
        exit 1
      fi
      BUMP_TYPE="$CUSTOM_VER"
      ;;
    *)
      BUMP_TYPE="patch"
      ;;
  esac
fi

case "$BUMP_TYPE" in
  patch)
    NEW_VERSION="$PATCH_PREVIEW"
    ;;
  minor)
    NEW_VERSION="$MINOR_PREVIEW"
    ;;
  major)
    NEW_VERSION="$MAJOR_PREVIEW"
    ;;
  *)
    NEW_VERSION="$BUMP_TYPE"
    ;;
esac

TAG_NAME="${TAG_PREFIX}${NEW_VERSION}"

echo -e "\n----------------------------------------"
echo -e "🎯 Target      : ${YELLOW}$(echo "$TARGET" | tr '[:lower:]' '[:upper:]')${NC}"
echo -e "📦 Versi Baru  : ${GREEN}${NEW_VERSION}${NC}"
echo -e "🏷️  Git Tag     : ${MAGENTA}${TAG_NAME}${NC}"
echo -e "----------------------------------------\n"

# 5. Check if tag already exists in local git
if git tag -l "$TAG_NAME" | grep -qx "$TAG_NAME"; then
  echo -e "${RED}❌ Tag \"$TAG_NAME\" sudah ada di lokal git! Gunakan versi lain.${NC}"
  exit 1
fi

# 6. Check uncommitted changes
GIT_STATUS=$(git status --porcelain)
if [ -n "$GIT_STATUS" ]; then
  echo -e "${YELLOW}⚠️  Perhatian: Ada perubahan file yang belum di-commit:${NC}"
  echo "$GIT_STATUS"
  read -p "Lanjutkan proses release? (y/N): " PROCEED
  PROCEED=$(echo "$PROCEED" | tr '[:upper:]' '[:lower:]' | xargs)
  if [ "$PROCEED" != "y" ]; then
    echo "❌ Release dibatalkan."
    exit 0
  fi
fi

# 7. Final confirmation
read -p "Yakin ingin merilis ${TAG_NAME}? (Y/n): " CONFIRM
CONFIRM=$(echo "$CONFIRM" | tr '[:upper:]' '[:lower:]' | xargs)
if [ -n "$CONFIRM" ] && [ "$CONFIRM" != "y" ]; then
  echo "❌ Release dibatalkan."
  exit 0
fi

# 8. Update VERSION file & commit
echo -e "\n📝 Mengupdate file VERSION..."
echo "$NEW_VERSION" > "$VERSION_FILE"
echo "  ✓ Updated VERSION to $NEW_VERSION"

COMMIT_MSG="chore(release): ${TAG_NAME}"
git add "$VERSION_FILE"
if git diff --cached --name-only | grep -qx "$VERSION_FILE"; then
  git commit -m "$COMMIT_MSG"
  echo "  ✓ Commit created: \"$COMMIT_MSG\""
else
  git commit --allow-empty -m "$COMMIT_MSG"
  echo "  ✓ Commit created: \"$COMMIT_MSG\""
fi

git tag -a "$TAG_NAME" -m "Release $TAG_NAME"
echo "  ✓ Tag created: $TAG_NAME"

# 9. Push to GitHub
read -p $'\n🚀 Push commit dan tag ke GitHub origin sekarang? (Y/n): ' PUSH_ANS
PUSH_ANS=$(echo "$PUSH_ANS" | tr '[:upper:]' '[:lower:]' | xargs)
if [ -z "$PUSH_ANS" ] || [ "$PUSH_ANS" = "y" ]; then
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
  echo -e "\nPushing to origin/${CURRENT_BRANCH}..."
  git push origin "$CURRENT_BRANCH"
  echo "Pushing tag ${TAG_NAME} to origin..."
  git push origin "$TAG_NAME"
  echo -e "\n${GREEN}🎉 Berhasil! Workflow GitHub Actions untuk $(echo "$TARGET" | tr '[:lower:]' '[:upper:]') otomatis berjalan.${NC}"
else
  echo -e "\nℹ️  Kamu bisa push manual kapan saja dengan:"
  echo "   git push origin && git push origin ${TAG_NAME}"
fi

