## When To Use

- When you need to show alert messages to users.
- When requiring users to interact with the application, but without jumping to a new page and interrupting the user's workflow, you can use Modal to create a new floating layer over the current page to get user feedback or display information.

### Initialize to Ensure the Use of a Unique Dialog in Quasar

```js
// ./boot/smartEnginEntrance.js
import { BtDialog } from '@bytetrade/ui';
import { Dialog } from 'quasar';

export default boot(async ({ app }) => {
  BtDialog.init(Dialog);
});
```

## API

### BtDialog.method()

- `BtDialog.show`

The items listed above are all functions, expecting a settings object as parameter. The properties of the object are follows:

API properties extends from Quasar [QDialogOptions](http://www.quasarchs.com/quasar-plugins/dialog)

| Property | Description | Type | Default | Version |
| --- | --- | --- | --- | --- |
| platform | Web or Mobile | string | web |  |
| title | title | string | Alert |  |
| message | message | string |  |  |
| icon | icon | string |  |  |
| ok | Confirmed events at the bottom | string \ number | true |  |
| cancel | Cancel event at the bottom | string \ number | false |  |
| okLoading | Shown when data is being sent | string \ boolean | false |  |
| okStyle | Confirm the style of the button | Object |  |  |
| prompt | Input Box Configuration | Object |  |  |

#### prompt

| Property    | Description   | Type    | Default | Version |
| ----------- | ------------- | ------- | ------- | ------- |
| model       | Initial value | string  |         |         |
| isValid     |               | boolean |         |         |
| type        | text          | string  |         |         |
| name        | Input label   | string  |         |         |
| placeholder | Hint text     | string  |         |         |

## Examples

### MessageDialog

```ts
BtDialog.show({
  title: 'Remove',
  message: 'Removing desc',
  okStyle: {
    background: '#FF7733',
    color: '#ffffff'
  },
  okText: 'Confirm',
  cancel: true
})
  .then(res => {
    if (res) {
      console.log('click ok');
      rssStore.removeFeed(feed.url);
    } else {
      console.log('click cancel');
    }
  })
  .catch(err => {
    console.log('click error', err);
  });
```

### EditDialog

```ts
import { BtDialog } from '@bytetrade/ui';
import DialogView from './views/DialogView.vue';

BtDialog.show({
  component: DialogView,
  title: 'Edit',
  platform: 'mobile',
  okStyle: {
    background: '#ff6600',
    color: '#919191'
  },
  cancel: true,
  prompt: {
    model: '',
    isValid: val => val.length > 2, // << here is the magic
    type: 'text', // optional
    name: 'View Namdsahkdas',
    placeholder: ''
  }
})
  .then(res => {
    if (res) {
      console.log('click ok', res);
    } else {
      console.log('click cancel', res);
    }
  })
  .catch(err => {
    console.log('click ok', err);
  });
```
