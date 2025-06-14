#!/bin/bash

# Build script for local development

echo "Installing Node.js dependencies..."
npm install

echo "Copying Alpine.js to static directory..."
mkdir -p ./static/js
cp node_modules/alpinejs/dist/cdn.min.js static/js/alpine.min.js

echo "Building Tailwind CSS..."
npm run build-css

echo "Build complete! You can now run your Go application or build the Docker image."
