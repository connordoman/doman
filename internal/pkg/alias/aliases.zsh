# DO NOT EDIT

# This script will load aliases from the `aliases` dir into your shell session

setopt EXTENDED_GLOB NULL_GLOB

SCRIPT_DIR="$(dirname "$(realpath "$0")")"

for f in "$SCRIPT_DIR/aliases/"*.zsh(Nr); do
    source "$f"
done
