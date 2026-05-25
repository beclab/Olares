## When To Use

- Display a notification message globally.

### Initialize to Ensure the Use of a Unique Notify in Quasar

```js
// ./boot/smartEnginEntrance.js
import { BtNotify } from '@bytetrade/ui';
import { Notify } from 'quasar';

export default boot(async ({ app }) => {
  BtNotify.init(Notify);
});
```

## API

### BtNotify.method()

- `BtNotify.show`
- `BtNotify.hide`

The items listed above are all functions, expecting a settings object as parameter. The properties of the object are follows:

API properties extends from Quasar [QDialogOptions](http://www.quasarchs.com/quasar-plugins/notify#notify-api)

| Property | Description | Type | Default | Version |
| --- | --- | --- | --- | --- |
| type | registerType | string | NotifyDefinedType |  |
| closeTimeout | Whether to turn off automatic disappearance | boolean | false |  |
| message | message | string |  |  |

#### NotifyDefinedType

| Property     | Description | Type   | Default     | Version |
| ------------ | ----------- | ------ | ----------- | ------- |
| DEFAULT      |             | string | bt-default  | 0.0.70  |
| LOGIN_NOTIFY |             | string | bt-login    | 0.0.70  |
| POSITIVE     |             | string | bt-positive | 0.0.70  |
| SUCCESS      |             | string | bt-success  | 0.0.70  |
| FAILED       |             | string | bt-failed   | 0.0.70  |
| WARNING      |             | string | bt-warning  | 0.0.70  |
| LOADING      |             | string | bt-loading  | 0.0.70  |
| PROGRESS     |             | string | bt-progress | 0.0.71  |

## Examples

```ts
import { BtNotify } from '@bytetrade/ui';
BtNotify.show({
  type: 'bt-success',
  message: 'This is a "info" type notification.',
  closeTimeout: true
});

BtNotify.hide();
```
