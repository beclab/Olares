#!/bin/bash

# Check if pngquant and jpegoptim are installed
if ! command -v pngquant &> /dev/null; then
    echo "pngquant is not installed. Please install it first."
    exit 1
fi

if ! command -v jpegoptim &> /dev/null; then
    echo "jpegoptim is not installed. Please install it first."
    exit 1
fi

# Create base compress directory if it doesn't exist
mkdir -p compress

# Function to process a single file
process_file() {
    local source_file="$1"
    # Remove 'public/' prefix and prepend 'compress/'
    local target_file="compress/${source_file#public/}"
    # Get the directory path of the target file
    local target_dir="$(dirname "$target_file")"
    
    # Create target directory if it doesn't exist
    mkdir -p "$target_dir"
    
    # Copy file first to preserve original
    cp "$source_file" "$target_file"
    
    # Process based on file extension
    local ext=$(echo "${source_file##*.}" | tr '[:upper:]' '[:lower:]')
    
    # Process based on file extension
    case "$ext" in
        png)
            echo "Compressing PNG: $source_file"
            pngquant --quality=90-95 --skip-if-larger --force --output "$target_file" "$target_file"
            ;;
        jpg|jpeg)
            echo "Compressing JPEG: $source_file"
            jpegoptim --max=95 --preserve --strip-none "$target_file"
            ;;
    esac
}

# Find and process all PNG and JPEG files
find public -type f \( -iname "*.png" -o -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.svg" \) | while read -r file; do
    process_file "$file"
done

echo "Image compression completed. Compressed images are in the 'compress' directory." 