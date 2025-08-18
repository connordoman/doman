# DO NOT EDIT

# This script will load aliases from the `aliases` dir into your shell session

setopt EXTENDED_GLOB NULL_GLOB

SCRIPT_DIR="$(dirname "$(realpath "$0")")"
echo "Zsh Aliases Loader ($SCRIPT_DIR)"

for f in "$SCRIPT_DIR/aliases/"*.zsh(Nr); do
    echo "Loading alias from $f"
    source "$f"
done
