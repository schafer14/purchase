Feature: Payment system

  A user should be able to pay for stuff and get
  good visibility on the status of the payment 
  and the tracking of the products.

  Scenario: Completed payment
    Given 1 order with a total payment of $32.00 USD
    And a payment is received for $32.00 USD
    When payment 1 is used for order 1
    Then payment 1 should be consumed
    And order 1 should be marked as paid-in-full
    And order 1 should have $0.00 USD remaining

  Scenario: Insufficent payment
    Given 1 order with a total payment of $32.00 USD
    And a payment is received for $31.00 USD
    When payment 1 is used for order 1
    Then payment 1 should be consumed
    And order 1 should be marked as paid-in-part
    And order 1 should have $1.00 USD remaining

  Scenario: Double spend
    Given 2 order with a total payment of $32.00 USD
    And a payment is received for $32.00 USD
    And payment 1 is used for order 1
    When payment 1 is used for order 2
    Then payment 1 should be consumed
    And order 2 should be marked as pending-payment
    And order 2 should have $32.00 USD remaining
    And order 1 should be marked as paid-in-full
    And order 1 should have $0.00 USD remaining
  
  Scenario: Completing paid-in-part purchase
    Given 1 order with a total payment of $64.00 USD
    And a payment is received for $32.00 USD
    And payment 1 is used for order 1
    And a payment is received for $32.00 USD
    When payment 2 is used for order 1
    Then payment 1 should be consumed
    And payment 2 should be consumed
    And order 1 should be marked as paid-in-full
    And order 1 should have $0.00 USD remaining

 Scenario: Over paid 
    Given 1 order with a total payment of $25.00 USD
    And a payment is received for $32.00 USD
    When payment 1 is used for order 1
    Then payment 1 should be consumed
    And order 1 should be marked as over-paid
    And order 1 should have $-7.00 USD remaining
