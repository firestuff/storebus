import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('replace success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo', num: 5});
	assert.equal(create.text, 'foo');

	const get1 = await tc.client.getTestType(create.id);
	assert.equal(get1.text, 'foo');
	assert.equal(get1.num, 5);

	const replace = await tc.client.replaceTestType(create.id, {text: 'bar'});
	assert.equal(replace.text, 'bar');

	const get2 = await tc.client.getTestType(create.id);
	assert.equal(get2.text, 'bar');
	assert.equal(get2.num, 0);
});
