---
outline: [2, 3]
description: Set up Plane on Olares for self-hosted project management. Organize work with modules, cycles, multiple views, and team collaboration.
head:
  - - meta
    - name: keywords
      content: Olares, Plane, project management, self-hosted, kanban, Gantt chart, team collaboration, task tracking
app_version: "1.0.9"
doc_version: "1.0"
doc_updated: "2026-03-30"
---

# Manage projects with Plane

Plane is an open-source project management platform that combines visual boards with agile task workflows. It helps teams plan sprints, track tasks, and collaborate on documents in one place.

Self-hosting Plane on Olares keeps all your project data under your control.

## Learning objectives

In this guide, you will learn how to:
- Install Plane and set up a workspace
- Configure email notifications via SMTP
- Invite and manage team members
- Create projects and organize work with modules, cycles, and work items
- Use different views to track progress

## Install Plane

1. Open Market and search for "Plane".
   <!-- ![Plane](/images/manual/use-cases/plane.png#bordered) -->
2. Click **Get**, then **Install**, and wait for installation to complete.

## Set up Plane

### Register and create a workspace

1. Open Plane from Launchpad and click **Get Started** to open the registration page.
2. Fill in your name, company email, and company name, then set a password and click **Continue**.
   <!-- ![Register for Plane](/images/manual/use-cases/plane-register.png#bordered) -->
3. Click **Create Workspace** in the bottom-right corner. Give your workspace a name, set the team size, and click **Create**.
   <!-- ![Create workspace](/images/manual/use-cases/plane-create-workspace.png#bordered) -->
4. Enter your display name and role (for example, `Developer`), then click **Continue** to enter the workspace.
5. Plane requires you to sign in after creating a workspace. Use the email and password you registered with.

### Configure email notifications

To enable email notifications for your team, configure SMTP in Plane's admin panel.

:::info Admin access required
Only the instance admin (the first registered user) can access god-mode.
:::

1. In your browser address bar, append `/god-mode` to your Plane URL to enter the admin panel. For example, if your Plane URL is `https://abc123.alice123.olares.com`, navigate to `https://abc123.alice123.olares.com/god-mode`.
2. Click **Email** in the left sidebar to open the SMTP configuration page.
3. Enter your SMTP server address, port, and the sender email address.
4. Enter the email account credentials and click **Save**.
5. Click **Send test email** to verify the configuration.

<!-- ![Configure SMTP](/images/manual/use-cases/plane-smtp-config.png#bordered) -->

## Manage your team

1. Click the workspace name in the top-left corner and select **Invite Members** to open workspace settings.
2. In the **Members** section, click **Add Member**.
3. Enter a team member's email address. You can invite multiple members at once. Once they accept the invitation, they join the workspace automatically.

<!-- ![Invite team members](/images/manual/use-cases/plane-invite-members.png#bordered) -->

:::tip
Invited members can check their **Workspace Invites** to view and accept pending invitations.
:::

## Use Plane

### Create a project

1. In your workspace, click **Get Started** to create a new project.
2. Name the project, add a description and icon, set project visibility, and assign a lead.

<!-- ![Create project](/images/manual/use-cases/plane-create-project.png#bordered) -->

### Break down work with modules

Modules let you group related tasks into logical units, such as feature areas or deliverables.

1. Expand the project in the sidebar and click **Modules**.
2. Click **Create Module**.
3. Enter a module name and description. Set the date range, status label, lead, and members.
4. Click **Create Module** to save.

<!-- ![Create module](/images/manual/use-cases/plane-create-module.png#bordered) -->

### Plan sprints with cycles

Cycles represent time-boxed iterations (sprints) for your team.

1. Expand the project in the sidebar and click **Cycles**.
2. Click **Set Your First Cycle**.
3. Enter a cycle name and description, then set the start and end dates.
4. Click **Create Cycle** to save.

<!-- ![Create cycle](/images/manual/use-cases/plane-create-cycle.png#bordered) -->

### Create and assign work items

1. Expand the project in the sidebar and click **Work Items**.
2. Click **Create** to add a new work item. Enter a title and description, set the status and priority, assign it to a team member, and link it to a cycle and module.
3. Click the work item to open its detail page, where you can attach files.
4. Click **Add Sub-Work Item** in the detail page to break the task into smaller pieces.

<!-- ![Create work items](/images/manual/use-cases/plane-create-work-item.png#bordered) -->

### Switch between views

Plane offers multiple ways to visualize your work. Click the view icons in the toolbar above your work items to switch between them.

| View | Best for | Key advantage |
|:-----|:---------|:--------------|
| Kanban | Daily standups and task tracking | Drag-and-drop status changes, swimlane grouping by module or assignee |
| Calendar | Planning deadlines | Automatically highlights overdue items |
| Spreadsheet | Bulk-editing task properties | Excel-like interface for quick updates to priority and assignee |
| Gantt chart | Managing cross-module dependencies | Visualizes task blocking risks across teams |

### Collaborate on documents

Plane includes a built-in document editor for team knowledge sharing.

1. Open the **Pages** section in your project to create a new document. Use the rich text editor to add headings, links, and code blocks. Team members can edit documents collaboratively.
2. Link documents to work items by attaching the document URL in a work item's detail page.
3. Export documents locally for offline access.

<!-- ![Document collaboration](/images/manual/use-cases/plane-documents.png#bordered) -->

## Learn more

- [Plane documentation](https://docs.plane.so/): Official guides for Plane features, workflows, and administration.
