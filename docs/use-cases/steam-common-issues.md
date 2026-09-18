---
outline: [2, 3]
description: "Troubleshoot common Steam Headless issues on Olares, including package persistence and Black Myth: Wukong Frame Generation."
head:
  - - meta
    - name: keywords
      content: Olares, Steam Headless, common issues, Flatpak, apt, Black Myth Wukong, DLSS Frame Generation, troubleshooting
app_version: "1.0.46"
doc_version: "1.1"
doc_updated: "2026-09-18"
---

# Steam Headless common issues

Find solutions to common Steam Headless problems on Olares.

## Packages installed with `apt` disappear after Steam Headless restarts

Packages installed with `apt` are written to the container's root filesystem. Steam Headless recreates this filesystem when the app restarts, is redeployed, or is upgraded. As a result, packages installed manually with `apt` are not retained.

To keep additional packages across Steam Headless restarts and updates, use Flatpak instead of `apt` when a Flatpak package is available. In Steam Headless 1.0.43 and later, Flatpak applications, runtimes, and user data are stored in persistent app storage.

To install a package with Flatpak:

1. Open Control Hub and go to **Browse** > **steamheadless**.
2. Expand **Deployments** > **steamheadless**, then open the running Pod.
3. Under **Containers**, click the Terminal icon next to **steam-headless**.
4. Run the Flatpak installation command in the container shell.

For more information about Pods and containers, see [Manage containers](../manual/olares/controlhub/manage-container.md).

## Frame Generation is unavailable in Black Myth: Wukong

When you run Black Myth: Wukong through Proton, **Frame Generation** might be unavailable in the graphics settings. Steam Headless 1.0.46 and later includes a script that configures DirectX 12, DLSS, and hardware-accelerated GPU scheduling for the game's Proton environment.

1. Install Black Myth: Wukong in Steam.
2. Launch the game once to create its Proton environment, then quit the game completely.
3. Open Control Hub and go to **Browse** > **steamheadless**.
4. Expand **Deployments** > **steamheadless**, then open the running Pod.
5. Under **Containers**, click the Terminal icon next to **steam-headless**.
6. Run the following command:

   ```bash
   /usr/bin/fix-wukong-frame-gen.sh
   ```

7. Check that the command finishes with the following message:

   ```plain
   [OK] Done. Launch Black Myth: Wukong and enable Frame Generation in the graphics menu.
   ```

8. Launch the game, open its graphics settings, and enable **Frame Generation**.

Run the script again after reinstalling or updating the game. If the script reports that the game is running, quit the game completely before retrying.
