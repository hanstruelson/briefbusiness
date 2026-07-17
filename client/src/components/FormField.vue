<template>
  <div class="form-field">
    <label
      v-if="label"
      :for="name"
      class="form-label">
      {{ label }}
      <span v-if="required" class="required" aria-hidden="true">*</span>
    </label>

    <div class="form-input-wrapper">
      <span
        v-if="errorMsg"
        class="field-error"
        role="alert"
        aria-live="polite">
        {{ errorMsg }}
      </span>

      <input
        ref="inputRef"
        :name="name"
        :type="inputType"
        :placeholder="placeholder"
        :value="modelValue"
        :disabled="disabled"
        class="form-input"
        :class="{ 'input-invalid': hasError }"
        @input="onInput"
        @blur="onBlur"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, defineProps, defineEmits } from 'vue'


const props = defineProps({
  name: { type: String, required: true },
  label: { type: String, default: '' },
  required: { type: Boolean, default: false },
  placeholder: { type: String, default: '' },
  inputType: { type: String, default: 'text' },
  modelValue: { type: String },
  errorMsg: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits([
  ['update:modelValue'],
  ['validation-error-clear']
])

const inputRef = ref(null)

const hasError = computed(() => !!props.errorMsg)

function onInput(event) {
  emit('update:modelValue', event.target.value)
  if (props.errorMsg) {
    emit('validation-error-clear', props.name)
  }
}

function onBlur() {
  if (props.errorMsg && props.modelValue.trim()) {
    emit('validation-error-clear', props.name)
  }
}
</script>

<style scoped>
.form-field {
  position: relative;
  margin-bottom: 18px;
}

.form-label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 6px;
  letter-spacing: 0.3px;
}

.required {
  color: #e53e3e;
  margin-left: 4px;
}

.form-input-wrapper {
  position: relative;
}

.field-error {
  display: block;
  font-size: 0.78rem;
  color: #e53e3e;
  margin-top: 4px;
  padding-left: 2px;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.95rem;
  color: #2d3748;
  background: #fafbfc;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
}

.form-input:disabled {
  background: #f0f2f5;
  cursor: not-allowed;
}

.input-invalid {
  border-color: #e53e3e;
  box-shadow: 0 0 0 3px rgba(229, 62, 62, 0.1);
}
</style>
