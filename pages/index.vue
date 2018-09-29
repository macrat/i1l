<style scoped>
main {
	min-width: 80%;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
}
img {
	margin-bottom: 2em;
}
form {
	width: 50em;
	display: flex;
	align-items: center;
	justify-content: center;
}
@media screen and (max-width: 50em) {
	#logo {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	main {
		display: block;
		padding: 0 1em;
	}
	form {
		display: block;
		width: auto;
		text-align: right;
	}
	#card {
		margin-top: 20%;
	}
}

#result {
	text-align: center;
	font-size: 200%;
}
</style>

<template>
	<main style="background-image: ">
		<div id=logo>
			<img src=~/assets/logo.svg>
		</div>

		<md-card id=card>
			<md-card-content>
				<form @submit.prevent=onsubmit>
					<md-field>
						<md-input ref=input type=url v-model=url placeholder=http:// required />
					</md-field>
					<md-button class="md-raised md-primary" type=submit>SHORTEN</md-button>
				</form>

				<md-field v-if="shorten != null">
					<md-input ref=result id=result @click=copy :value=shorten readonly @copy.native=copied />
				</md-field>
			</md-card-content>
		</md-card>

		<md-dialog-alert
			:md-active="errorMessage != null"
			@md-closed="errorMessage = null"
			md-title="failed to shorten!"
			md-content=errorMessage
			/>

		<md-snackbar :md-active.sync=copyMessage :md-duration=3000>
			<span>copied to clipboard</span>
			<md-button class="md-primary" @click="copyMessage = false">OK</md-button>
		</md-snackbar>
	</main>
</template>

<script>
import Vue from 'vue';
import {MdButton, MdField, MdCard, MdSnackbar} from 'vue-material/dist/components';
import 'vue-material/dist/vue-material.min.css';
import 'vue-material/dist/theme/default.css';
import axios from 'axios';

Vue.use(MdButton);
Vue.use(MdField);
Vue.use(MdCard);
Vue.use(MdSnackbar);

export default {
	data() {
		return {
			url: 'http://',
			shorten: null,
			errorMessage: null,
			copyMessage: false,
		};
	},
	mounted() {
		this.$refs.input.$el.focus();
		this.$refs.input.$el.selectionStart = 10;
	},
	methods: {
		onsubmit() {
			axios({
				method: 'post',
				url: '/',
				headers: {'Content-Type': 'text/plain'},
				data: this.url,
				responseType: 'json',
			}).then(x => {
				if (x.data.error != null) {
					this.errorMessage = e;
				} else {
					this.shorten = x.data.info.short_url;
				}
			}).catch(e => this.errorMessage = e)
		},
		copy() {
			this.$refs.result.$el.select();
			document.execCommand('copy');
		},
		copied() {
			this.copyMessage = true;
		},
	},
}
</script>
