---
outline: [2, 3]
description: Learn how to use the Clone feature in Olares to run multiple instances of the same application with independent configurations.
---

# Clone applications

You can install multiple instances of the same application, which is also known as "clone". This is useful for scenarios requiring different configurations or versions, such as running multiple Windows virtual machines.

## Clone an application

1. Confirm that the target application supports cloning:    
    
    a. Open **Market** from Dock or Launchpad, and click the application to open the details page.

    b. Check the **Required permissions** section. Only applications with the **Multi-Instance** property support cloning.
2. Click **My Olares** from the left sidebar. The list of installed applications will appear on the right.
3. Locate the target application, click the drop-down arrow next to the **Open** button, and then click **Clone**.

    ![Clone applications](/images/manual/olares/clone-apps.png#bordered)

4. In the **Clone app** window, configure the instance details:
    
   - **App Name**: Enter a unique name to identify the application in the **Market** and **My Olares** list.
   - **Display Name**: Set the name for the desktop shortcut. You can name each icon individually if the application has multiple entry points.
5. Click **Confirm** to proceed.
6. In the **Configure Environment Variables** window, complete the corresponding setup, and then click **Confirm**. 

    In the installed applications list, the newly created application will appear with a "Clone" tag next to it.

    ![Clone applications result](/images/manual/olares/clone-apps-result.png#bordered)
