var pg_info = function() {
	// Sort tables
	var sort = function(headers) {
		$(headers || 'table.sort th').on('click', function(e) {
			var th       = $(this),
				num_sort = th.is('.n'),
				col      = th.index(),
				tbody    = th.closest('table').find('>tbody'),
				rows     = Array.from(tbody.find('>tr')),
				to_i     = (i) => parseInt(i.replace(/,/g, ''), 10),
				is_sort  = th.attr('data-sort') === '1'

			if (num_sort)
				rows.sort((a, b) => to_i(a.children[col].innerText) < to_i(b.children[col].innerText))
			else
				rows.sort((a, b) => a.children[col].innerText.localeCompare(b.children[col].innerText))
			if (is_sort)
				rows.reverse()

			tbody.html('').html(rows)
			th.closest('table').find('th').attr('data-sort', '0')
			th.attr('data-sort', is_sort ? '0' : '1')
		})
	}
	sort()

	// Collapse sections.
	$('h2').on('click', function(e) {
		var next = $(this).next()
		next.css('display', (next.css('display') === 'none' ? 'block' : 'none'))
	})

	// Query explain
	$('#explain form').on('submit', function(e) {
		e.preventDefault()

		var form = $(this),
			ta   = form.find('textarea')

		jQuery.ajax({
			method: 'POST',
			url:    form.attr('action'),
			data:   form.serialize(),
			success: function(data) {
				form.after($('<pre class="e"></pre>').html(data).append('' +
					'<form action="https://explain.dalibo.com/new" method="POST" target="_blank">' +
						'<input type="hidden" name="plan"  value="' + data + '">' +
						'<input type="hidden" name="query" value="' + form.find('textarea').val() + '">' +
						'<button type="submit">PEV</button>' +
					'</form>'))
			}
		})
	})

	// Load table details
	$('.load-table').on('click', function(e) {
		e.preventDefault()

		var row = $(this).closest('tr')
		if (row.next().is('.table-detail'))
			return row.next().remove()

		jQuery.ajax({
			// TODO: use path prefix
			url: '/table/' + $(this).text(),
			success: function(data) {
				var nrow = $('<tr class="table-detail"><td colspan="10"></td></tr>')
				nrow.find('td').html(data)
				row.after(nrow)
				sort(nrow.find('table th'))
			},
		})
	})
}

pg_info()
