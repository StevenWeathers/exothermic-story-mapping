<script>
    import Modal from './Modal.svelte'
    import UserRegisterForm from './UserRegisterForm.svelte'

    export let notifications
    export let toggleCreate = () => {}
    export let handleCreate = () => {}
</script>

<Modal closeModal="{toggleCreate}">
    <UserRegisterForm handleSubmit="{handleCreate}" {notifications} />
</Modal>
