---
outline: [2, 3]
description: "Troubleshoot common Steam Headless issues on Olares, including package persistence and Black Myth: Wukong Frame Generation."
head:
  - - meta
    - name: keywords
      content: Olares, Steam Headless, common issues, Flatpak, apt, Black Myth Wukong, DLSS Frame Generation, troubleshooting
app_version: "1.0.49"
doc_version: "1.1"
doc_updated: "2026-09-20"
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

When you run Black Myth: Wukong through Proton, **Frame Generation** might be unavailable in the graphics settings. The latest Steam Headless release includes a script that configures DirectX 12, DLSS, and hardware-accelerated GPU scheduling for the game's Proton environment.

1. Open Market and update Steam Headless to the latest available version.
2. In the Steam Library, select Black Myth: Wukong. Wait for any download or file validation to finish, and make sure **Play** is available.
3. Click **Play** and wait until the game reaches the main menu. Then quit the game completely.
4. Open Control Hub and go to **Browse** > **steamheadless**.
5. Expand **Deployments** > **steamheadless**, then open the running Pod.
6. Under **Containers**, click the Terminal icon next to **steam-headless**.
7. Confirm that the fix script is available:

   ```bash
   command -v fix-wukong-frame-gen.sh
   ```

   The command should return:

   ```plain
   /usr/bin/fix-wukong-frame-gen.sh
   ```

   If the command returns no output, return to Market and confirm that Steam Headless is up to date before continuing.

8. Run the fix script:

   ```bash
   /usr/bin/fix-wukong-frame-gen.sh
   ```

9. Check that the output includes the following messages:

   ```plain
   [OK] GameUserSettings.ini: Dx12=1, Dlss=1
   [OK] Wrote registry HwSchMode=2 into system.reg
   [OK] Done. Launch Black Myth: Wukong and enable Frame Generation in the graphics menu.
   ```

10. Launch the game, open its graphics settings, and enable **Frame Generation**.

Run the script again after reinstalling or updating the game. If the script reports that the game is running, quit the game completely before retrying. Do not rely on the `-dx12` Steam launch option because the Steam client might remove it.
