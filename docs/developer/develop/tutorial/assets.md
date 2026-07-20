---
description: Learn how to prepare and upload icons, feature image, and promotional images for your Olares apps.
---
# Add icons, feature image, and promotional images

A great-looking app needs high-quality assets. This guide covers the specifications for your app's icon, feature image, and screenshots, and how to upload them to Olares.

## Asset specifications
Before uploading, ensure your images are in the correct format.

| Type                  | Format          | Max size    | Dimensions (px) | Description                                                                                                           |
|:----------------------|:----------------|:------------|:----------------|:----------------------------------------------------------------------------------------------------------------------|
| **App icon**          | PNG, WEBP       | 512 KB      | 256x256         | Your app's most common visual symbol, used on the Olares desktop and throughout the system.                           |
| **Feature image**     | JPEG, PNG, WEBP | 8 MB        | 1440x900        | Displayed on your app's page in **Market** > **My Olares**.                                                           |
| **Promotional image** | JPEG, PNG, WEBP | 8 MB (each) | 1440x900        | If you plan to submit your app to the public Market, you must upload at least two. You can upload a maximum of eight. |

## Upload and link assets

1. Navigate to the [Olares Market Image Hosting service](https://imghost.olares.com/).
2. Select the type of asset you are uploading (e.g., Icon).
3. Drag and drop your prepared file into the upload area, or click to select it.
4. Click the image thumbnail to make simple edits if necessary.
5. When you are ready, click **Upload**. 
6. After the upload, the service will provide a direct URL for the image. Click <span class="material-symbols-outlined">content_copy</span> to copy the URL to your clipboard.
7. Open your app project in **Studio**.
8. Paste the URL into the corresponding field in your `OlaresManifest.yaml` file.