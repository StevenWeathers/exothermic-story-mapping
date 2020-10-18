<script>
    import { _ } from '../i18n'

    export let avatarService = ''
    export let userId = ''
    export let avatar = ''
    export let width = 32
</script>

{#if avatarService === 'dicebear'}
    <img
        src="https://avatars.dicebear.com/api/{avatar}/{userId}.svg?w={width}"
        alt="{$_('avatarAltText')}" />
{:else if avatarService === 'gravatar'}
    <img
        src="https://gravatar.com/avatar/{userId}?s={width}&d={avatar}&r=g"
        alt="{$_('avatarAltText')}" />
{:else if avatarService === 'robohash'}
    <img
        src="https://robohash.org/{userId}.png?set={avatar}&size={width}x{width}"
        alt="{$_('avatarAltText')}" />
{:else if avatarService === 'govatar'}
    <img src="/avatar/{width}/{userId}/{avatar}" alt="{$_('avatarAltText')}" />
{:else if avatarService === 'goadorable'}
    <img src="/avatar/{width}/{userId}" alt="{$_('avatarAltText')}" />
{:else}
    <img
        src="https://api.adorable.io/avatars/{width}/{userId}.png"
        alt="{$_('avatarAltText')}" />
{/if}
