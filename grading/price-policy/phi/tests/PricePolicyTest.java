package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import lib.*;

public class PricePolicyTest {
    @Test public void defaultPricingUsesQuantityTimesUnitPrice() {
        PricePolicy policy=new DefaultPricePolicy();
        for(int quantity:new int[]{0,1,2,7,13})for(int price:new int[]{0,1,57,199})assertEquals(quantity*price,policy.calculateCost(quantity,price),0.0001);
    }
    @Test public void buyOneGetOneHandlesZeroOddAndEvenQuantities() {
        PricePolicy policy=new B1G1PricePolicy();
        for(int quantity=0;quantity<=15;quantity++)assertEquals("Charge for ceil(quantity/2) items",((quantity+1)/2)*137,policy.calculateCost(quantity,137),0.0001);
    }
    @Test public void discountHandlesFractionalPercentagesAndSetterChanges() {
        DiscountPricePolicy policy=new DiscountPricePolicy();
        for(double percentage:new double[]{0,12.5,25,99,100}) {
            policy.setPercentage(percentage);assertEquals(percentage,policy.getPercentage(),0.0001);assertEquals(7*137*(1-percentage/100),policy.calculateCost(7,137),0.0001);
        }
    }
    @Test public void orderLineDelegatesToAnUnfamiliarPolicyAndCanSwitchPolicies() {
        PricePolicy custom=new PricePolicy(){public String getBriefDescription(){return "Custom";}public double calculateCost(int quantity,int price){return quantity*price+17.25;}};
        OrderLine line=new OrderLine("Item",137,3,custom);assertSame(custom,line.getPolicy());assertEquals(428.25,line.getCost(),0.0001);
        line.setQuantity(5);line.setUnitPrice(91);assertEquals(472.25,line.getCost(),0.0001);
        line.setPolicy(new B1G1PricePolicy());assertEquals(273,line.getCost(),0.0001);
        line.setPolicy(new DiscountPricePolicy(50));assertEquals(227.5,line.getCost(),0.0001);
    }
}
