<script>
    import SolidButton from './SolidButton.svelte'
    import Modal from './Modal.svelte'
    import { _ } from '../i18n'

    export let handleDeleteAccount = () => {}
    export let toggleDeleteAccount = () => {}
</script>

<Modal closeModal={toggleDeleteAccount}>
    <div class="mb-4">
        <p class="font-bold text-xl text-red-600">
            Are you sure you want to delete your account, this
            cannot be undone.
        </p>
    </div>
    <div class="text-right">
        <button
            type="button"
            class="inline-block align-baseline font-bold text-sm
            text-blue-500 hover:text-blue-800 mr-4"
            on:click="{toggleDeleteAccount}">
            Cancel
        </button>
        <SolidButton onClick="{handleDeleteAccount}" color="red">
            Delete Account
        </SolidButton>
    </div>
</Modal>