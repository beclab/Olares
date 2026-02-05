---
outline: [2, 3] 
description: Optimize your app listing with images and icons in Olares Market.
---
# Promote your apps

High-quality visuals help your application stand out in Olares Market. This guide covers asset specifications and how to generate image URLs for your app listing.

## Application assets

Configure these assets in your [`OlaresManifest.yaml`](/developer/develop/package/manifest.md).

### Application icon

**Required**. Displayed on the Launchpad and in the Market list.
- **Location**: Configure the icon URL in the icon field under `metadata` or `entrances` in `OlaresManifest.yaml`.
- **Format**: PNG or WEBP
- **Dimensions**: 256 × 256 pixels
- **Size**: No larger than 512 KB

### Promote images

**Recommended**. Displayed on the application details page. We recommend uploading at least 2 images.
- **Location**: Configure image URLs in the `promoteImage` field under `spec` in `OlaresManifest.yaml`.
- **Format**: JPEG, PNG, or WEBP
- **Dimensions**: 1440 × 900 pixels
- **Size**: No larger than 8 MB per image
- **Limit**: Up to 8 images

### Featured image

**Optional**. Used for recommendations in Olares Market or displayed on the "My Olares" section.
- **Location**: Configure the image URL in the `featuredImage` field under `spec` in `OlaresManifest.yaml`.
- **Format**: JPEG, PNG, or WEBP
- **Dimensions**: 1440 × 900 pixels
- **Size**: No larger than 8 MB
- **Limit**: One image only

### Image hosting service 

You can also host images on your own server or use the Olares image hosting service:

1. Open [Olares Market image hosting](https://imghost.olares.com/).
2. Select the image type: **app icon**, **featured image** or **promotional image**.
3. Upload and preview the image.
4. Copy the generated URL and paste it into your `OlaresManifest.yaml`.