<template>
  <div class="signup-container">
    <div class="signup-card">
      <header class="signup-header">
        <h1>Create Account</h1>
        <p class="signup-subtitle">Join us today</p>
      </header>

      <form @submit.prevent="handleSubmit" class="signup-form">
        <!-- Email -->
        <form-field name="Email" label="Email Address" required v-model="d.formData.Email" type="email"
          placeholder="e.g. john@example.com" />

        <!-- Submit Button -->
        <button type="submit" class="btn-primary" :disabled="d.isLoading">
          <span class="btn-text">
            {{ d.isLoading ? 'Creating Account...' : 'Create Account' }}
          </span>
          <svg v-if="d.isLoading" class="btn-spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor"
            stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <path d="M4 12a8 8 0 018-8" />
          </svg>
        </button>

        <!-- Error Message -->
        <div v-if="d.errorMessage" class="error-message">
          {{ d.errorMessage }}
        </div>
      </form>

      <div class="signup-footer">
        <p>
          Already have an account?
          <a href="/login">Sign In</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, defineEmits } from 'vue';
import { useRouter } from 'vue-router';

const emit = defineEmits(['userRegistered']);
const router = useRouter();

const d = reactive({
  formData: {
    Email: 'test@test.com' + Math.floor(Math.random() * 100000),
  },
});

const isValidEmail = computed(() => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(d.formData.Email);
});
const hasErrors = computed(() => {
  return !d.formData.Email;
});

const handleSubmit = async () => {
  const errors = [];

  if (!d.formData.Email || !isValidEmail.value) {
    errors.push('Valid Email is required');
  }
  if (errors.length > 0) {
    d.errorMessage = errors.join(' ') || 'Please fix the errors above';
    return;
  }

  try {
    d.isLoading = true;
    d.errorMessage = '';

    const response = await fetch('/api/user/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(d.formData),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      d.errorMessage = errorData.message || 'Failed to create account';
      return;
    }
    const registerData = await response.json();
    localStorage.setItem('AccessToken', registerData.AccessToken);
    localStorage.setItem('UserId', registerData.UserId);

    emit('userRegistered', registerData);
    router.push('/');
    
    d.formData.Email = '';
    d.errorMessage = '';
  } catch (error) {
    d.errorMessage = 'Network error. Please try again.';
  } finally {
    d.isLoading = false;
  }
};
</script>

<style scoped>
.signup-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  margin: 0;
  padding: 20px;
}

.signup-card {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  width: 100%;
  max-width: 420px;
  overflow: hidden;
}

.signup-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  text-align: center;
  padding: 32px 24px 28px;
}

.signup-header h1 {
  font-size: 1.8rem;
  font-weight: 700;
  margin: 0;
  letter-spacing: -0.5px;
}

.signup-subtitle {
  margin-top: 8px;
  font-size: 0.95rem;
  opacity: 0.9;
}

.signup-form {
  padding: 24px;
}

.form-field {
  margin-bottom: 18px;
  position: relative;
}

.form-field label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 6px;
  letter-spacing: 0.3px;
}

.form-field input {
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

.form-field input:focus {
  outline: none;
  border-color: #667eea;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
}

.form-field input.invalid {
  border-color: #e53e3e;
  box-shadow: 0 0 0 3px rgba(229, 62, 62, 0.1);
}

.btn-primary {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  border: none;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn-spinner {
  width: 18px;
  height: 18px;
}

.error-message {
  background: #fef2f2;
  color: #e53e3e;
  padding: 12px 16px;
  border-radius: 10px;
  font-size: 0.85rem;
  margin-bottom: 16px;
  border: 1px solid #fadbd8;
  text-align: center;
}

.signup-footer {
  text-align: center;
  padding: 20px 24px;
  border-top: 1px solid #edf2f7;
}

.signup-footer p {
  color: #718096;
  font-size: 0.85rem;
  margin: 0;
}

.signup-footer a {
  color: #667eea;
  text-decoration: none;
  font-weight: 600;
  cursor: pointer;
}

.signup-footer a:hover {
  text-decoration: underline;
}
</style>
