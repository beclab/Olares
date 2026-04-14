---
outline: [2, 3]
description: Set up Plane on Olares for self-hosted project management. Organize work with modules, cycles, multiple views, and team collaboration.
head:
  - - meta
    - name: keywords
      content: Olares, Plane, project management, self-hosted, kanban, Gantt chart, team collaboration, task tracking
app_version: "1.0.9"
doc_version: "1.0"
doc_updated: "2026-04-14"
---

# Manage projects with Plane

Plane is an open-source project management platform that combines visual boards with agile task workflows. It helps teams plan sprints, track tasks, and collaborate on documents in one place.

Self-hosting Plane on Olares keeps all your project data under your control.

## Learning objectives

In this guide, you will learn how to:
- Install Plane and the background services it relies on.
- Set up a workspace and bring your team onboard.
- Run a project by categorizing work, scheduling sprints, and assigning tasks.
- Gain insights into the project progress using different visual layouts.

## Install Plane and required dependencies

Before installing Plane, you must first install its dependencies: RabbitMQ (V4.0.0 or later) and MinIO (V1.0.0 or later).

1. Open Market and search for "RabbitMQ".

   ![RabbitMQ](/images/manual/use-cases/rabbitmq.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.
3. Search for "MinIO" and install it.
   
   ![MinIO](/images/manual/use-cases/minio.png#bordered)

4. Search for "Plane" and install it.
   
   ![Plane](/images/manual/use-cases/plane.png#bordered)

## Create your workspace

After installation, register your account and create a workspace where all your projects will reside.

1. Open Plane from the Launchpad, and then click **Get started** on the welcome page.
2. On the **Setup your Plane Instance** page, fill in the required information, and then click **Continue**.

   ![Register for Plane](/images/manual/use-cases/plane-register.png#bordered){width=60%}

3. Click **Create workspace** in the bottom-right corner.

   ![Create workspace](/images/manual/use-cases/plane-create-workspace.png#bordered)

4. Specify your workspace name, select the team size, and then click **Create workspace**.
5. Click the newly created workspace, and then log in with the email and password you just set up.
6. Complete your profile settings, and then click **Continue**. You enter the new workspace.

## Invite your team

Invite team members into the workspace so they can collaborate with you.

1. Click your workspace name in the top-left corner, and then select **Invite members**.
2. On the **Members** page, click **Add member**.
3. Enter the email address of the team member, and then select the role to assign to the member.
4. To invite multiple members at once, click **Add more**.

   ![Invite team members](/images/manual/use-cases/plane-invite-members.png#bordered){width=80%}

5. Click **Send invitations**. Invitees appear in the **Pending invites** panel.

   When the invited members accept the invitation, they join the workspace automatically.

   :::tip
   Invited members can accept the invitation by checking the **Workspace invites** section in their own Plane interface.
   :::

## Use Plane

To see how to use Plane to manage a multi-stage project, let’s walk through a sample scenario: your team needs to execute a "Product Page Revamp" to increase website conversion rates.

### Create a project

Start by creating a project for this specific initiative and adding the team members who will execute the work.

1. Open your workspace and create a new project:
   - For your first project, click **Get started** on the **Home** page.
   - For subsequent projects, select **Projects** from the left sidebar, and then click **Add Project**.
2. Define the core details of the initiative:
   - **Project name**: `Product Page Revamp`
   - **Project ID**: `WEB`
   - **Description**: `Improve UX/UI and messaging for the core product landing page`
3. Select an icon, set project visibility, and assign a lead.
4. Click **Create project**.

   ![Create project](/images/manual/use-cases/plane-create-project.png#bordered){width=70%}

5. Click **Open project**.
6. In the left sidebar, click the new project name, click <span class="material-symbols-outlined">more_horiz</span>, and then click **Settings**.
7. In the left sidebar, under the project name, select **Members**.
8. Click **Add member**, select the members contributing to this project and their roles, and then click **Add members**.
9. Click **Back to workspace** in the top-left corner.

### Categorize work

To keep a large project organized, it helps to group related tasks into logical categories. In Plane, these categories are called "Modules".

In this scenario, we create three modules: "Visual assets", "Copywriting", and "Technical SEO".

1. From the left sidebar, click the new project to expand it, and then click **Modules**.
2. Click **Build your first module** or **Add Module**.
3. Define the core details of the module:
   - **Title**: `Visual assets`
   - **Description**: `Focus on photography, iconography, and UI design elements`
4. Set the date range, status label, lead, and members.
5. Click **Create Module**.

   ![Create module](/images/manual/use-cases/plane-create-module.png#bordered){width=70%}

6. Repeat these steps to create the other two modules.

### Schedule work into sprints

Instead of tackling everything at once, break your timeline down into focused, time-boxed periods known as sprints. In Plane, sprints are called "Cycles".

In this scenario, we will create two cycles to show the transition from planning to execution.

1. In the left sidebar, click **Cycles**.
2. Click **Set your first cycle** or **Add cycle**.
3. Define the core details of the phase:
   - **Title**: `Phase 1: Discovery`
   - **Description**: `Research, wireframe, and define the core value proposition; the goal is to finalize the skeleton of the new product page`
4. Select the start and end dates.
5. Click **Create cycle**.

   ![Create cycle](/images/manual/use-cases/plane-create-cycle.png#bordered){width=70%}

6. Repeat these steps to create the cycle for Phase 2:
   - **Title**: `Phase 2: Execution`
   - **Description**: `Hi-Fi UI design, final copy production, and SEO auditing; the goal is to complete final visual assets and prepare for development`

### Create and assign work items

Now that your structure is in place, detail the specific actions required to complete the revamp. Assign these action items to your team, set their priority, and map them to the categories and phases you just built.

In this scenario, we will create the following task items for the project.

| Task title | Module | Cycle | Priority |
|:---|:---|:---|:---|
| Conduct UX audit | Technical SEO | Phase 1: Discovery | High |
| Draft eye-catching headlines | Copywriting | Phase 1: Discovery | Urgent |
| Create Low-Fi sketches | Visual Assets | Phase 1: Discovery | Medium |
| Design final UI mockups | Visual Assets | Phase 2: Execution | High |
| Write meta descriptions | Technical SEO | Phase 2: Execution | Medium |

1. In the left sidebar, click **Work items**.
2. Click **Create your first work item** or **Add work item**.
3. Define the core details of the task:
   - **Title**: `Conduct UX audit`
   - **Description**: `Review the current homepage for friction points, and focus on mobile navigation and Add to Cart button visibility`.
4. Set the status and priority, assign it to a team member, set the date range, and link it to a cycle and module.

   ![Create work items](/images/manual/use-cases/plane-create-work-item.png#bordered){width=70%}

5. Click **Save**.
6. Repeat these steps to create work items for the remaining tasks in this scenario.
7. To add more context for a work item, click it to open the details page, where you can attach files, add sub-work items, or leave comments.

   ![Work item details page](/images/manual/use-cases/plane-work-item-details.png#bordered)

### Track progress

Depending on whether you are running a daily standup meeting, checking upcoming deadlines, or looking for scheduling conflicts, you will need to look at your data differently.

Use the layout icons in the upper-right corner of the **Work items** page to switch views and get the insights you need.

![Layout views](/images/manual/use-cases/plane-layouts.png#bordered)

The following layouts are available:
- **List Layout**: Groups tasks into collapsible sections (like Todo and In Progress) so you can quickly see where everything stands at a glance.
- **Board Layout**: Displays tasks as Kanban cards, allowing you to easily drag and drop work from column to column as it progresses. 
- **Calendar Layout**: Plots your tasks on a traditional monthly calendar grid so you can see exactly when deliverables are due.
- **Table Layout**: Provides a spreadsheet-style interface with distinct columns, making it easy to review and update priorities, assignees, and labels in bulk. 
- **Timeline Layout**: Maps out task durations as horizontal bars in Gantt style to help you pace the project and spot overlapping work.

### Draft and share project resources

Share knowledge across your team. Keep everyone aligned by writing and storing project resources right next to the work itself.

1. In the left sidebar, click **Pages**.
2. Click **Create your first page** or **Add page**.
3. Enter a title for the new document, such as `Revamp strategy for homepage 2026`.
4. Click **Create Page**.
5. Use the editor to draft your document collaboratively with your team.

   ![Document collaboration](/images/manual/use-cases/plane-documents.png#bordered) 

6. To connect a document to a work item, click <span class="material-symbols-outlined">link_2</span> in the upper-right corner to copy the document URL, and then paste it into a work item's description so the assignee has the context needed.
7. To save a local copy of the document, click <span class="material-symbols-outlined">more_horiz</span> in the upper-right corner, and then click **Export**.

## Learn more

- [Official Plane documentation](https://docs.plane.so/)
