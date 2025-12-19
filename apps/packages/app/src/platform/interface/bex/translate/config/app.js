/*
 * @Author: yyh 24493052+yongheng2016@users.noreply.github.com
 * @Date: 2025-05-13 20:35:31
 * @LastEditors: yyh 24493052+yongheng2016@users.noreply.github.com
 * @LastEditTime: 2025-05-14 15:34:24
 * @FilePath: /kiss-translator/src/config/app.js
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import { REACT_APP_NAME } from '../env';

export const APP_NAME = REACT_APP_NAME.trim().split(/\s+/).join('-');
export const APP_LCNAME = APP_NAME.toLowerCase();
