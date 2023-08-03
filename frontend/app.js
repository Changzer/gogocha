// app.js

const apiUrl = 'http://localhost:8080';

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('orderForm').addEventListener('submit', submitOrder);
    fetchProductList();
});

document.getElementById('name').addEventListener('input', (event) => {
    order.customer.name = event.target.value;
    updateOrderSummary(); // Call this to update the display
});

const order = {
    customer: {},
    products: [],
};

function fetchProductList() {
    fetch(apiUrl + '/products')
        .then(response => response.json())
        .then(products => {
            displayProductList(products);
        })
        .catch(error => {
            console.error('Error fetching products:', error);
            alert('Failed to fetch products');
        });
}

function displayProductList(products) {
    const productListContainer = document.getElementById('productList');
    productListContainer.innerHTML = ''; // Clear the previous content

    products.forEach(product => {
        const productContainer = document.createElement('div');
        productContainer.className = 'product-container';

        const productNameElement = document.createElement('span');
        productNameElement.textContent = product.product_name;
        productContainer.appendChild(productNameElement);

        const priceElement = document.createElement('span');
        const price = parseFloat(product.price);
        if (!isNaN(price)) {
            priceElement.textContent = '$' + price.toFixed(2);
        } else {
            priceElement.textContent = 'Price not available';
        }
        productContainer.appendChild(priceElement);

        const quantityInput = document.createElement('input');
        quantityInput.type = 'number';
        quantityInput.min = '1';
        quantityInput.value = '1';
        productContainer.appendChild(quantityInput);

        const addToOrderButton = document.createElement('button');
        addToOrderButton.textContent = 'Add to Order';
        addToOrderButton.addEventListener('click', () => {
            addProductToOrder(product, parseInt(quantityInput.value));
        });
        productContainer.appendChild(addToOrderButton);

        productListContainer.appendChild(productContainer);
    });
}

function addProductToOrder(product, quantity) {
    let productInOrder = order.products.find(p => p.product_id === product.product_id);

    if (!productInOrder) {
        productInOrder = {
            product_id: product.product_id,
            product_name: product.product_name,
            price: parseFloat(product.price),
            quantity: 0,
        };
        order.products.push(productInOrder);
    }

    productInOrder.quantity += quantity;

    updateOrderSummary();
}

function removeProductFromOrder(productId) {
    const productIndex = order.products.findIndex(p => p.product_id === productId);
    if (productIndex > -1) {
        order.products.splice(productIndex, 1);
    }
    updateOrderSummary(); // Function to re-render the order summary
}

function updateOrderSummary() {
    const summaryElement = document.getElementById('orderSummary');
    summaryElement.innerHTML = ''; // Clear current summary

    // Display Customer Name
    const customerInfo = document.createElement('div');
    customerInfo.textContent = `Customer Name: ${order.customer.name || 'N/A'}`;
    summaryElement.appendChild(customerInfo);

    let totalAmount = 0;
    order.products.forEach(product => {
        const itemElement = document.createElement('div');
        itemElement.textContent = `${product.product_name}: ${product.quantity} x ${product.price}`;

        const removeButton = document.createElement('button');
        removeButton.textContent = 'Remove';
        removeButton.addEventListener('click', () => removeProductFromOrder(product.product_id));

        itemElement.appendChild(removeButton);
        summaryElement.appendChild(itemElement);

        // Calculate total amount
        totalAmount += product.price * product.quantity;
    });

    // Display Total Amount
    const totalAmountElement = document.createElement('div');
    totalAmountElement.textContent = `Total Amount: $${totalAmount.toFixed(2)}`;
    summaryElement.appendChild(totalAmountElement);
}



function submitOrder(event) {
    event.preventDefault();

    // Calculate total amount
    let totalAmount = 0;
    for (const productId in order.products) {
        const product = order.products[productId];
        totalAmount += product.price * product.quantity;
    }

    // Prepare order data
    const orderData = {
        customer: {
            name: order.customer.name,
            cpf: order.customer.cpf,
            email: order.customer.email,
        },
        products: Object.values(order.products), // Convert products to an array
        totalAmount: totalAmount.toFixed(2), // Include total amount
        order_date: new Date().toISOString(), // Include the current date/time
    };

    // Log the order data to verify
    console.log('Order data before sending:', orderData);

    // Make the POST request to create the order
    fetch(apiUrl + '/orders', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(orderData),
    })
        .then(response => {
            if (!response.ok) {
                // Handle error response
                return response.text().then(text => {
                    throw new Error(text);
                });
            }
            return response.json();
        })
        .then(data => {
            console.log('Order created:', data);
            alert('Order created successfully! Order ID: ' + data.order_id);
            resetOrder(); // You may want to reset the order after submission
        })
        .catch(error => {
            console.error('Error:', error.message);
            alert('Failed to create order: ' + error.message);
        });
}


function resetOrder() {
    order.customer = {};
    order.products = [];
    updateOrderSummary();
}
