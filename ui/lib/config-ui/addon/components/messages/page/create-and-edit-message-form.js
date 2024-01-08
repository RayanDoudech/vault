/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: BUSL-1.1
 */

import Component from '@glimmer/component';
import { tracked } from '@glimmer/tracking';
import { task } from 'ember-concurrency';
import errorMessage from 'vault/utils/error-message';
import { inject as service } from '@ember/service';

/**
 * @module Page::CreateAndEditMessageForm
 * Page::CreateAndEditMessageForm components are used to display create and edit message form fields.
 * @example
 * ```js
 * <Page::CreateAndEditMessageForm @message={{this.message}}  />
 * ```
 * @param {model} message - message model to pass to form components
 */

export default class MessagesList extends Component {
  @service router;
  @service store;
  @service flashMessages;

  @tracked errorBanner = '';
  @tracked modelValidations;
  @tracked invalidFormMessage;
  @tracked showMessagePreviewModal = false;

  willDestroy() {
    super.willDestroy();
    const noTeardown = this.store && !this.store.isDestroying;
    const { model } = this;
    if (noTeardown && model && model.get('isDirty') && !model.isDestroyed && !model.isDestroying) {
      model.rollbackAttributes();
    }
  }

  @task
  *save(event) {
    event.preventDefault();
    try {
      const { isValid, state, invalidFormMessage } = this.args.message.validate();
      this.modelValidations = isValid ? null : state;
      this.invalidFormAlert = invalidFormMessage;

      if (isValid) {
        const { isNew } = this.args.message;
        const { id, title } = yield this.args.message.save();
        this.flashMessages.success(`Successfully ${isNew ? 'created' : 'updated'} ${title} message.`);
        this.store.clearDataset('config-ui/message');
        this.router.transitionTo('vault.cluster.config-ui.messages.message.details', id);
      }
    } catch (error) {
      this.errorBanner = errorMessage(error);
      this.invalidFormAlert = 'There was an error submitting this form.';
    }
  }
}